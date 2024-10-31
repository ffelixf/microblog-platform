// internal/repository/tweet_repository_test.go
package repository

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ffelixf/microblog-platform/internal/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTweetTestDB(t *testing.T) (*mongo.Client, func()) {
	ctx := context.Background()
	uri := "mongodb://admin:adminpassword@localhost:27017/test_db?authSource=admin"

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		t.Fatalf("Error connecting to MongoDB: %v", err)
		return nil, nil
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		t.Fatalf("Error connecting to MongoDB: %v", err)
		return nil, nil
	}

	t.Log("Successfully connected to MongoDB")

	// Crear índices necesarios
	collection := client.Database("test_db").Collection("tweets")
	_, err = collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}},
		},
	})
	if err != nil {
		t.Fatalf("Error creating indexes: %v", err)
		return nil, nil
	}

	cleanup := func() {
		if err := client.Database("test_db").Collection("tweets").Drop(ctx); err != nil {
			t.Logf("Error dropping tweets collection: %v", err)
		}
		if err := client.Database("test_db").Collection("users").Drop(ctx); err != nil {
			t.Logf("Error dropping users collection: %v", err)
		}
		if err := client.Disconnect(ctx); err != nil {
			t.Logf("Error disconnecting from MongoDB: %v", err)
		}
	}

	return client, cleanup
}

func createTestUserForTweets(t *testing.T, client *mongo.Client) primitive.ObjectID {
	ctx := context.Background()
	userID := primitive.NewObjectID()
	_, err := client.Database("test_db").Collection("users").InsertOne(ctx, bson.M{
		"_id":             userID,
		"username":        "testuser",
		"email":           "test@example.com",
		"following":       []string{},
		"followers_count": 0,
		"created_at":      time.Now(),
		"updated_at":      time.Now(),
	})
	assert.NoError(t, err)
	return userID
}

func TestTweetRepository_Create(t *testing.T) {
	client, cleanup := setupTweetTestDB(t)
	defer cleanup()

	repo := NewTweetRepository(client, "test_db")
	ctx := context.Background()
	userID := createTestUserForTweets(t, client)

	t.Run("successful tweet creation", func(t *testing.T) {
		tweet := &models.Tweet{
			UserID:  userID,
			Content: "Test tweet content",
		}

		err := repo.Create(ctx, tweet)
		assert.NoError(t, err)
		assert.NotEmpty(t, tweet.ID)
		assert.NotZero(t, tweet.CreatedAt)
		assert.Equal(t, userID, tweet.UserID)
	})

	t.Run("tweet without content", func(t *testing.T) {
		tweet := &models.Tweet{
			UserID:  userID,
			Content: "",
		}

		err := repo.Create(ctx, tweet)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no puede estar vacío")
	})

	t.Run("tweet exceeding max length", func(t *testing.T) {
		longContent := strings.Repeat("a", 281)
		tweet := &models.Tweet{
			UserID:  userID,
			Content: longContent,
		}

		err := repo.Create(ctx, tweet)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no puede exceder los 280 caracteres")
	})

	t.Run("tweet with valid max length", func(t *testing.T) {
		content := strings.Repeat("a", 280)
		tweet := &models.Tweet{
			UserID:  userID,
			Content: content,
		}

		err := repo.Create(ctx, tweet)
		assert.NoError(t, err)
		assert.NotEmpty(t, tweet.ID)
	})

	t.Run("tweet with non-existent user", func(t *testing.T) {
		nonExistentUserID := primitive.NewObjectID()
		tweet := &models.Tweet{
			UserID:  nonExistentUserID,
			Content: "Test content",
		}

		err := repo.Create(ctx, tweet)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "usuario especificado no existe")
	})

	t.Run("tweet with zero user ID", func(t *testing.T) {
		tweet := &models.Tweet{
			UserID:  primitive.ObjectID{},
			Content: "Test content",
		}

		err := repo.Create(ctx, tweet)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID de usuario es requerido")
	})
}

func TestTweetRepository_GetByUserID(t *testing.T) {
	client, cleanup := setupTweetTestDB(t)
	defer cleanup()

	repo := NewTweetRepository(client, "test_db")
	ctx := context.Background()

	t.Run("get user tweets", func(t *testing.T) {
		userID := createTestUserForTweets(t, client)

		// Crear varios tweets
		for i := 0; i < 3; i++ {
			tweet := &models.Tweet{
				UserID:  userID,
				Content: fmt.Sprintf("Test tweet %d", i),
			}
			err := repo.Create(ctx, tweet)
			assert.NoError(t, err)
		}

		tweets, err := repo.GetByUserID(ctx, userID.Hex())
		assert.NoError(t, err)
		assert.Len(t, tweets, 3)

		// Verificar orden por fecha
		for i := 1; i < len(tweets); i++ {
			assert.True(t, tweets[i-1].CreatedAt.After(tweets[i].CreatedAt) ||
				tweets[i-1].CreatedAt.Equal(tweets[i].CreatedAt))
		}
	})

	t.Run("user with no tweets", func(t *testing.T) {
		emptyUserID := createTestUserForTweets(t, client)
		tweets, err := repo.GetByUserID(ctx, emptyUserID.Hex())
		assert.NoError(t, err)
		assert.Empty(t, tweets)
	})

	t.Run("invalid user ID format", func(t *testing.T) {
		tweets, err := repo.GetByUserID(ctx, "invalid-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID de usuario inválido")
		assert.Nil(t, tweets)
	})

	t.Run("ordered by creation date", func(t *testing.T) {
		userID := createTestUserForTweets(t, client)

		// Crear tweets con diferentes fechas
		for i := 0; i < 3; i++ {
			tweet := &models.Tweet{
				UserID:  userID,
				Content: fmt.Sprintf("Tweet %d", i),
			}
			err := repo.Create(ctx, tweet)
			assert.NoError(t, err)
			time.Sleep(time.Millisecond * 100) // Asegurar diferentes timestamps
		}

		tweets, err := repo.GetByUserID(ctx, userID.Hex())
		assert.NoError(t, err)
		assert.Len(t, tweets, 3)

		// Verificar orden descendente por fecha
		for i := 1; i < len(tweets); i++ {
			assert.True(t, tweets[i-1].CreatedAt.After(tweets[i].CreatedAt))
		}
	})
}

func TestTweetRepository_GetTimeline(t *testing.T) {
	client, cleanup := setupTweetTestDB(t)
	defer cleanup()

	repo := NewTweetRepository(client, "test_db")
	ctx := context.Background()

	t.Run("get timeline with tweets", func(t *testing.T) {
		// Crear usuarios
		userID := createTestUserForTweets(t, client)
		followedID1 := createTestUserForTweets(t, client)
		followedID2 := createTestUserForTweets(t, client)

		// Establecer relaciones de following
		_, err := client.Database("test_db").Collection("users").UpdateOne(
			ctx,
			bson.M{"_id": userID},
			bson.M{"$set": bson.M{"following": []string{followedID1.Hex(), followedID2.Hex()}}},
		)
		assert.NoError(t, err)

		// Crear tweets para los usuarios seguidos
		tweet1 := &models.Tweet{
			UserID:  followedID1,
			Content: "Tweet from followed user 1",
		}
		err = repo.Create(ctx, tweet1)
		assert.NoError(t, err)

		tweet2 := &models.Tweet{
			UserID:  followedID2,
			Content: "Tweet from followed user 2",
		}
		err = repo.Create(ctx, tweet2)
		assert.NoError(t, err)

		// Obtener timeline
		tweets, err := repo.GetTimeline(ctx, userID.Hex(), 1, 10)
		assert.NoError(t, err)
		assert.Len(t, tweets, 2)

		// Verificar orden cronológico inverso
		assert.True(t, tweets[0].CreatedAt.After(tweets[1].CreatedAt) ||
			tweets[0].CreatedAt.Equal(tweets[1].CreatedAt))
	})

	t.Run("timeline pagination", func(t *testing.T) {
		// Crear usuario y seguidor
		userID := createTestUserForTweets(t, client)
		followedID := createTestUserForTweets(t, client)

		// Establecer relación de following
		_, err := client.Database("test_db").Collection("users").UpdateOne(
			ctx,
			bson.M{"_id": userID},
			bson.M{"$set": bson.M{"following": []string{followedID.Hex()}}},
		)
		assert.NoError(t, err)

		// Crear varios tweets
		for i := 0; i < 15; i++ {
			tweet := &models.Tweet{
				UserID:  followedID,
				Content: fmt.Sprintf("Paginated tweet %d", i),
			}
			err := repo.Create(ctx, tweet)
			assert.NoError(t, err)
			time.Sleep(time.Millisecond * 10) // Asegurar diferentes timestamps
		}

		// Probar primera página
		page1, err := repo.GetTimeline(ctx, userID.Hex(), 1, 10)
		assert.NoError(t, err)
		assert.Len(t, page1, 10)

		// Probar segunda página
		page2, err := repo.GetTimeline(ctx, userID.Hex(), 2, 10)
		assert.NoError(t, err)
		assert.Len(t, page2, 5)

		// Verificar que no hay tweets duplicados
		for _, tweet1 := range page1 {
			for _, tweet2 := range page2 {
				assert.NotEqual(t, tweet1.ID, tweet2.ID)
			}
		}
	})

	t.Run("empty timeline", func(t *testing.T) {
		userID := createTestUserForTweets(t, client)
		tweets, err := repo.GetTimeline(ctx, userID.Hex(), 1, 10)
		assert.NoError(t, err)
		assert.Empty(t, tweets)
	})

	// Enhanced Timeline Tests
	t.Run("invalid page number", func(t *testing.T) {
		userID := createTestUserForTweets(t, client)
		tweets, err := repo.GetTimeline(ctx, userID.Hex(), 0, 10)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "número de página inválido")
		assert.Nil(t, tweets)
	})

	t.Run("invalid page size", func(t *testing.T) {
		userID := createTestUserForTweets(t, client)
		tweets, err := repo.GetTimeline(ctx, userID.Hex(), 1, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tamaño de página inválido")
		assert.Nil(t, tweets)
	})

	t.Run("max page size exceeded", func(t *testing.T) {
		userID := createTestUserForTweets(t, client)
		tweets, err := repo.GetTimeline(ctx, userID.Hex(), 1, 101)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tamaño de página máximo excedido")
		assert.Nil(t, tweets)
	})

	t.Run("invalid user ID format", func(t *testing.T) {
		tweets, err := repo.GetTimeline(ctx, "invalid-id", 1, 10)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID de usuario inválido")
		assert.Nil(t, tweets)
	})

	t.Run("non-existent user", func(t *testing.T) {
		nonExistentID := primitive.NewObjectID()
		tweets, err := repo.GetTimeline(ctx, nonExistentID.Hex(), 1, 10)
		assert.NoError(t, err)
		assert.Empty(t, tweets)
	})
}
