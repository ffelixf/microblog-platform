// cmd/api/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ffelixf/microblog-platform/internal/handlers"
	"github.com/ffelixf/microblog-platform/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func connectDB() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		return nil, fmt.Errorf("MONGODB_URI no está configurado en .env")
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	log.Println("✅ Conexión exitosa a MongoDB")
	return client, nil
}

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":    "ok",
		"message":   "Server is running",
		"timestamp": time.Now(),
	})
}

func dbHealthCheck(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		dbStatus := "ok"
		dbError := ""

		if client == nil {
			dbStatus = "error"
			dbError = "MongoDB client not initialized"
		} else {
			err := client.Ping(ctx, nil)
			if err != nil {
				dbStatus = "error"
				dbError = err.Error()
			}

			_, err = client.ListDatabaseNames(ctx, bson.M{})
			if err != nil {
				dbStatus = "error"
				dbError = err.Error()
			}
		}

		response := gin.H{
			"status":    "ok",
			"timestamp": time.Now(),
			"database": gin.H{
				"status": dbStatus,
				"type":   "mongodb",
				"uri":    os.Getenv("MONGODB_URI"),
			},
		}

		if dbError != "" {
			response["database"].(gin.H)["error"] = dbError
			c.JSON(500, response)
			return
		}

		c.JSON(200, response)
	}
}

func main() {
	// Configurar logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Conectar a MongoDB
	mongoClient, err := connectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer mongoClient.Disconnect(context.Background())

	// Inicializar repositorios
	userRepo := repository.NewUserRepository(mongoClient, os.Getenv("MONGODB_DATABASE"))
	tweetRepo := repository.NewTweetRepository(mongoClient, os.Getenv("MONGODB_DATABASE"))

	// Inicializar handlers
	userHandler := handlers.NewUserHandler(userRepo)
	tweetHandler := handlers.NewTweetHandler(tweetRepo)

	// Configurar router
	r := gin.Default()

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Registrar rutas
	handlers.RegisterUserRoutes(r, userHandler)
	handlers.RegisterTweetRoutes(r, tweetHandler)

	// Health checks
	r.GET("/health", healthCheck)
	r.GET("/health/db", dbHealthCheck(mongoClient))

	// Obtener puerto
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("📝 Using default port: %s", port)
	}

	// Iniciar servidor
	log.Printf("🚀 Server starting on http://localhost:%s", port)
	log.Printf("💡 Health endpoint: http://localhost:%s/health", port)
	log.Printf("💡 DB Health endpoint: http://localhost:%s/health/db", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}
