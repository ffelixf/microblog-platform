// internal/repository/tweet_repository.go
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ffelixf/microblog-platform/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TweetRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
}

func NewTweetRepository(client *mongo.Client, dbName string) *TweetRepository {
	db := client.Database(dbName)
	collection := db.Collection("tweets")
	return &TweetRepository{
		collection: collection,
		db:         db,
	}
}

func (r *TweetRepository) Create(ctx context.Context, tweet *models.Tweet) error {
	// Validar que existe el usuario
	if tweet.UserID.IsZero() {
		return fmt.Errorf("el ID de usuario es requerido")
	}

	// Validar contenido del tweet
	if tweet.Content == "" {
		return fmt.Errorf("el contenido del tweet no puede estar vacío")
	}

	// Validar longitud máxima
	if len(tweet.Content) > 280 {
		return fmt.Errorf("el contenido del tweet no puede exceder los 280 caracteres")
	}

	// Validar que el usuario existe
	err := r.db.Collection("users").FindOne(ctx, bson.M{"_id": tweet.UserID}).Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("el usuario especificado no existe")
		}
		return fmt.Errorf("error al verificar usuario: %v", err)
	}

	tweet.CreatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, tweet)
	if err != nil {
		return fmt.Errorf("error al crear tweet: %v", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		tweet.ID = oid
	}
	return nil
}

func (r *TweetRepository) GetByUserID(ctx context.Context, userID string) ([]models.Tweet, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("ID de usuario inválido: %v", err)
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": objectID}, opts)
	if err != nil {
		return nil, fmt.Errorf("error al buscar tweets: %v", err)
	}
	defer cursor.Close(ctx)

	var tweets []models.Tweet
	if err = cursor.All(ctx, &tweets); err != nil {
		return nil, fmt.Errorf("error al decodificar tweets: %v", err)
	}

	return tweets, nil
}

func (r *TweetRepository) GetTimeline(ctx context.Context, userID string, page, limit int) ([]models.Tweet, error) {
	// Validar parámetros de paginación
	if page < 1 {
		return nil, fmt.Errorf("número de página inválido: debe ser mayor a 0")
	}
	if limit < 1 {
		return nil, fmt.Errorf("tamaño de página inválido: debe ser mayor a 0")
	}
	if limit > 100 {
		return nil, fmt.Errorf("tamaño de página máximo excedido: máximo 100")
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("ID de usuario inválido: %v", err)
	}

	// Obtener la lista de usuarios seguidos
	var user models.User
	err = r.db.Collection("users").FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []models.Tweet{}, nil
		}
		return nil, fmt.Errorf("error al obtener usuario: %v", err)
	}

	// Preparar lista de IDs para la consulta
	followingIDs := []primitive.ObjectID{objectID} // Incluir tweets propios
	for _, id := range user.Following {
		if objID, err := primitive.ObjectIDFromHex(id); err == nil {
			followingIDs = append(followingIDs, objID)
		}
	}

	// Configurar opciones de búsqueda
	skip := (page - 1) * limit
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	// Buscar tweets
	cursor, err := r.collection.Find(ctx,
		bson.M{"user_id": bson.M{"$in": followingIDs}},
		opts,
	)
	if err != nil {
		return nil, fmt.Errorf("error al obtener tweets: %v", err)
	}
	defer cursor.Close(ctx)

	var tweets []models.Tweet
	if err = cursor.All(ctx, &tweets); err != nil {
		return nil, fmt.Errorf("error al decodificar tweets: %v", err)
	}

	return tweets, nil
}
