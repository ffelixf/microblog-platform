// internal/models/user.go
package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
    ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Username      string            `bson:"username" json:"username"`
    Email         string            `bson:"email" json:"email"`
    CreatedAt     time.Time         `bson:"created_at" json:"created_at"`
    UpdatedAt     time.Time         `bson:"updated_at" json:"updated_at"`
    Following     []string          `bson:"following" json:"following"`
    FollowersCount int              `bson:"followers_count" json:"followers_count"`
}

// internal/models/tweet.go
type Tweet struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
    Content   string            `bson:"content" json:"content"`
    CreatedAt time.Time         `bson:"created_at" json:"created_at"`
}