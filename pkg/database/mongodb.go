// pkg/database/mongodb.go
package database

import (
    "context"
    "fmt"
    "time"
    "os"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func ConnectMongoDB() (*mongo.Client, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Obtener URI de MongoDB desde variables de entorno
    uri := os.Getenv("MONGODB_URI")
    if uri == "" {
        uri = "mongodb://microblog_user:microblog_password@localhost:27017/microblog"
    }

    // Configurar cliente MongoDB
    clientOptions := options.Client().ApplyURI(uri)
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
    }

    // Verificar conexión
    err = client.Ping(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
    }

    Client = client
    return client, nil
}

// GetCollection retorna una colección específica
func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
    database := os.Getenv("MONGODB_DATABASE")
    if database == "" {
        database = "microblog"
    }
    return client.Database(database).Collection(collectionName)
}