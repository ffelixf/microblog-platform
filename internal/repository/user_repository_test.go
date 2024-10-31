// internal/repository/user_repository_test.go
package repository

import (
	"context"
	"testing"

	"github.com/ffelixf/microblog-platform/internal/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Configuración de prueba para MongoDB
func setupTestDB(t *testing.T) (*mongo.Client, func()) {
	ctx := context.Background()

	// URI de conexión con credenciales
	uri := "mongodb://admin:adminpassword@localhost:27017/test_db?authSource=admin"

	clientOpts := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		t.Fatalf("Error connecting to MongoDB: %v", err)
		return nil, nil
	}

	// Verificar la conexión
	err = client.Ping(ctx, nil)
	if err != nil {
		t.Fatalf("Error connecting to MongoDB: %v", err)
		return nil, nil
	}

	t.Log("Successfully connected to MongoDB")

	// Crear índices únicos
	collection := client.Database("test_db").Collection("users")
	_, err = collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		t.Fatalf("Error creating indexes: %v", err)
		return nil, nil
	}

	// Función de limpieza
	cleanup := func() {
		// Limpiar la colección de prueba
		if err := client.Database("test_db").Collection("users").Drop(ctx); err != nil {
			t.Logf("Error dropping test collection: %v", err)
		}
		// Desconectar el cliente
		if err := client.Disconnect(ctx); err != nil {
			t.Logf("Error disconnecting from MongoDB: %v", err)
		}
	}

	return client, cleanup
}

// Helper función para crear un usuario de prueba
func createTestUser(t *testing.T, repo *UserRepository, username, email string) *models.User {
	user := &models.User{
		Username: username,
		Email:    email,
	}
	err := repo.Create(context.Background(), user)
	assert.NoError(t, err, "Error creating test user")
	return user
}

func TestUserRepository_Create(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(client, "test_db")
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		user := &models.User{
			Username: "testuser1",
			Email:    "test1@example.com",
		}

		err := repo.Create(ctx, user)
		assert.NoError(t, err)
		assert.NotEmpty(t, user.ID)
		assert.NotZero(t, user.CreatedAt)
		assert.NotZero(t, user.UpdatedAt)
		assert.Empty(t, user.Following)
		assert.Equal(t, 0, user.FollowersCount)
	})

	t.Run("duplicate username", func(t *testing.T) {
		// Limpiar la base de datos antes de la prueba
		err := client.Database("test_db").Collection("users").Drop(ctx)
		assert.NoError(t, err)

		// Crear índices nuevamente
		collection := client.Database("test_db").Collection("users")
		_, err = collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
			{
				Keys:    bson.D{{Key: "username", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys:    bson.D{{Key: "email", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
		})
		assert.NoError(t, err)

		// Crear el primer usuario
		user1 := &models.User{
			Username: "sameuser",
			Email:    "test1@example.com",
		}
		err = repo.Create(ctx, user1)
		assert.NoError(t, err)

		// Intentar crear un segundo usuario con el mismo username
		user2 := &models.User{
			Username: "sameuser",
			Email:    "test2@example.com",
		}
		err = repo.Create(ctx, user2)
		assert.Error(t, err, "Expected error for duplicate username")
		assert.Contains(t, err.Error(), "duplicate key error")
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(client, "test_db")
	ctx := context.Background()

	t.Run("get existing user", func(t *testing.T) {
		createdUser := createTestUser(t, repo, "getuser", "get@example.com")

		foundUser, err := repo.GetByID(ctx, createdUser.ID.Hex())
		assert.NoError(t, err)
		assert.NotNil(t, foundUser)
		assert.Equal(t, createdUser.ID, foundUser.ID)
		assert.Equal(t, createdUser.Username, foundUser.Username)
		assert.Equal(t, createdUser.Email, foundUser.Email)
	})

	t.Run("user not found", func(t *testing.T) {
		nonExistentID := primitive.NewObjectID()
		user, err := repo.GetByID(ctx, nonExistentID.Hex())
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserRepository_FollowUser(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(client, "test_db")
	ctx := context.Background()

	t.Run("successful follow", func(t *testing.T) {
		// Crear dos usuarios para la prueba
		follower := createTestUser(t, repo, "follower", "follower@example.com")
		followee := createTestUser(t, repo, "followee", "followee@example.com")

		// Seguir al usuario
		err := repo.FollowUser(ctx, follower.ID.Hex(), followee.ID.Hex())
		assert.NoError(t, err)

		// Verificar que el follower está siguiendo al followee
		updatedFollower, err := repo.GetByID(ctx, follower.ID.Hex())
		assert.NoError(t, err)
		assert.Contains(t, updatedFollower.Following, followee.ID.Hex())

		// Verificar que el contador de seguidores del followee se incrementó
		updatedFollowee, err := repo.GetByID(ctx, followee.ID.Hex())
		assert.NoError(t, err)
		assert.Equal(t, 1, updatedFollowee.FollowersCount)
	})

	t.Run("cannot follow self", func(t *testing.T) {
		user := createTestUser(t, repo, "selffollow", "self@example.com")
		err := repo.FollowUser(ctx, user.ID.Hex(), user.ID.Hex())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no puedes seguirte a ti mismo")
	})

	t.Run("follow non-existent user", func(t *testing.T) {
		user := createTestUser(t, repo, "follower2", "follower2@example.com")
		nonExistentID := primitive.NewObjectID()
		err := repo.FollowUser(ctx, user.ID.Hex(), nonExistentID.Hex())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "usuario objetivo no encontrado")
	})
}

func TestUserRepository_UnfollowUser(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(client, "test_db")
	ctx := context.Background()

	t.Run("successful unfollow", func(t *testing.T) {
		// Crear usuarios para la prueba
		follower := createTestUser(t, repo, "follower3", "follower3@example.com")
		followee := createTestUser(t, repo, "followee3", "followee3@example.com")

		// Primero hacer que el usuario siga al otro
		err := repo.FollowUser(ctx, follower.ID.Hex(), followee.ID.Hex())
		assert.NoError(t, err)

		// Verificar que el following se estableció correctamente
		updatedFollower, err := repo.GetByID(ctx, follower.ID.Hex())
		assert.NoError(t, err)
		assert.Contains(t, updatedFollower.Following, followee.ID.Hex())

		// Dejar de seguir
		err = repo.UnfollowUser(ctx, follower.ID.Hex(), followee.ID.Hex())
		assert.NoError(t, err)

		// Verificar que ya no lo sigue
		updatedFollower, err = repo.GetByID(ctx, follower.ID.Hex())
		assert.NoError(t, err)
		assert.NotContains(t, updatedFollower.Following, followee.ID.Hex())

		// Verificar que el contador de seguidores disminuyó
		updatedFollowee, err := repo.GetByID(ctx, followee.ID.Hex())
		assert.NoError(t, err)
		assert.Equal(t, 0, updatedFollowee.FollowersCount)
	})

	t.Run("unfollow non-followed user", func(t *testing.T) {
		user1 := createTestUser(t, repo, "user1", "user1@example.com")
		user2 := createTestUser(t, repo, "user2", "user2@example.com")

		err := repo.UnfollowUser(ctx, user1.ID.Hex(), user2.ID.Hex())
		assert.NoError(t, err) // No debería dar error, simplemente no hace nada
	})
}

func TestUserRepository_GetFollowing(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewUserRepository(client, "test_db")
	ctx := context.Background()

	t.Run("get following list", func(t *testing.T) {
		// Crear usuarios para la prueba
		follower := createTestUser(t, repo, "follower4", "follower4@example.com")
		followee1 := createTestUser(t, repo, "followee4", "followee4@example.com")
		followee2 := createTestUser(t, repo, "followee5", "followee5@example.com")

		// Seguir a dos usuarios
		err := repo.FollowUser(ctx, follower.ID.Hex(), followee1.ID.Hex())
		assert.NoError(t, err)
		err = repo.FollowUser(ctx, follower.ID.Hex(), followee2.ID.Hex())
		assert.NoError(t, err)

		// Obtener lista de usuarios seguidos
		following, err := repo.GetFollowing(ctx, follower.ID.Hex())
		assert.NoError(t, err)
		assert.Len(t, following, 2)

		// Verificar que los usuarios en la lista son correctos
		usernames := []string{followee1.Username, followee2.Username}
		for _, user := range following {
			assert.Contains(t, usernames, user.Username)
		}
	})

	t.Run("empty following list", func(t *testing.T) {
		user := createTestUser(t, repo, "loner", "loner@example.com")
		following, err := repo.GetFollowing(ctx, user.ID.Hex())
		assert.NoError(t, err)
		assert.Empty(t, following)
	})

	t.Run("non-existent user", func(t *testing.T) {
		nonExistentID := primitive.NewObjectID()
		following, err := repo.GetFollowing(ctx, nonExistentID.Hex())
		assert.Error(t, err, "Debería retornar error para usuario no existente")
		assert.Nil(t, following, "La lista de following debería ser nil para usuario no existente")
		assert.Contains(t, err.Error(), "usuario no encontrado",
			"El mensaje de error debería indicar que el usuario no fue encontrado")
	})
}
