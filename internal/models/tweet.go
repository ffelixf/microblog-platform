// internal/models/tweet.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Tweet struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id" binding:"required"`
	Content   string             `bson:"content" json:"content" binding:"required,max=280"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
