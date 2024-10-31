// internal/repository/user_repository.go
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ffelixf/microblog-platform/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(client *mongo.Client, dbName string) *UserRepository {
	collection := client.Database(dbName).Collection("users")
	return &UserRepository{
		collection: collection,
	}
}

// Método existente Create
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Following = make([]string, 0)
	user.FollowersCount = 0

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		user.ID = oid
	}
	return nil
}

// Método existente GetByID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Nuevo método FollowUser
func (r *UserRepository) FollowUser(ctx context.Context, userID, targetID string) error {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	targetObjID, err := primitive.ObjectIDFromHex(targetID)
	if err != nil {
		return err
	}

	// Verificar que el usuario objetivo existe
	var targetUser models.User
	err = r.collection.FindOne(ctx, bson.M{"_id": targetObjID}).Decode(&targetUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("usuario objetivo no encontrado")
		}
		return err
	}

	// Verificar que no se está siguiendo a sí mismo
	if userID == targetID {
		return errors.New("no puedes seguirte a ti mismo")
	}

	// Actualizar following del usuario
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userObjID},
		bson.M{"$addToSet": bson.M{"following": targetID}},
	)
	if err != nil {
		return err
	}

	// Incrementar followers_count del usuario objetivo
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": targetObjID},
		bson.M{"$inc": bson.M{"followers_count": 1}},
	)
	return err
}

// Nuevo método UnfollowUser
func (r *UserRepository) UnfollowUser(ctx context.Context, userID, targetID string) error {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	targetObjID, err := primitive.ObjectIDFromHex(targetID)
	if err != nil {
		return err
	}

	// Remover de la lista de following
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userObjID},
		bson.M{"$pull": bson.M{"following": targetID}},
	)
	if err != nil {
		return err
	}

	// Decrementar followers_count
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": targetObjID},
		bson.M{"$inc": bson.M{"followers_count": -1}},
	)
	return err
}

// Nuevo método GetFollowing
func (r *UserRepository) GetFollowing(ctx context.Context, userID string) ([]models.User, error) {
	// Convertir el ID a ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("ID de usuario inválido: %v", err)
	}

	// Buscar el usuario primero para verificar que existe
	var user models.User
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("usuario no encontrado")
		}
		return nil, err
	}

	// Si el usuario no sigue a nadie, retornar lista vacía
	if len(user.Following) == 0 {
		return []models.User{}, nil
	}

	// Convertir los IDs de following a ObjectIDs
	var followingObjIDs []primitive.ObjectID
	for _, id := range user.Following {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		followingObjIDs = append(followingObjIDs, objID)
	}

	// Si no hay IDs válidos, retornar lista vacía
	if len(followingObjIDs) == 0 {
		return []models.User{}, nil
	}

	// Buscar los usuarios que está siguiendo
	cursor, err := r.collection.Find(ctx, bson.M{
		"_id": bson.M{"$in": followingObjIDs},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var following []models.User
	if err = cursor.All(ctx, &following); err != nil {
		return nil, err
	}

	return following, nil
}

func (r *UserRepository) GetFollowers(ctx context.Context, userID string) ([]models.User, error) {
	// Buscar usuarios que tienen este userID en su array "following"
	cursor, err := r.collection.Find(ctx, bson.M{
		"following": userID,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var followers []models.User
	if err = cursor.All(ctx, &followers); err != nil {
		return nil, err
	}

	return followers, nil
}
