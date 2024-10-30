// internal/repository/user_repository.go
package repository

import (
	"context"
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
