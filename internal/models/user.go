// internal/models/user.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username       string             `bson:"username" json:"username" binding:"required"`
	Email          string             `bson:"email" json:"email" binding:"required,email"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
	Following      []string           `bson:"following" json:"following"`
	FollowersCount int                `bson:"followers_count" json:"followers_count"`
}
