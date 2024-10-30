// pkg/database/mongodb.go
package database

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Obtener URI de MongoDB desde variables de entorno
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("MONGODB_URI no está configurado en .env")
	}

	// Conectar a MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	// Verificar la conexión
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Conectado a MongoDB!")
	return client
}

// GetCollection obtiene una colección específica
func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	database := os.Getenv("MONGODB_DATABASE")
	if database == "" {
		database = "microblog"
	}
	return client.Database(database).Collection(collectionName)
}
