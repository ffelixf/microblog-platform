// cmd/api/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/ffelixf/microblog-platform/internal/handlers"
	"github.com/ffelixf/microblog-platform/internal/repository"
)

var mongoClient *mongo.Client

func connectDB() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Obtener URI de MongoDB
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		return nil, fmt.Errorf("MONGODB_URI no está configurado en .env")
	}

	// Conectar a MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("error conectando a MongoDB: %v", err)
	}

	// Verificar la conexión
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error haciendo ping a MongoDB: %v", err)
	}

	log.Println("✅ Conexión exitosa a MongoDB")
	return client, nil
}

func main() {
	// Configurar logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️  Warning: .env file not found: %v", err)
	}

	// Conectar a MongoDB
	var err error
	mongoClient, err = connectDB()
	if err != nil {
		log.Printf("❌ Error conectando a MongoDB: %v", err)
	} else {
		defer mongoClient.Disconnect(context.Background())
	}

	// Inicializar router
	r := gin.Default()

	// Ruta básica de health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"message":   "Server is running",
			"timestamp": time.Now(),
		})
	})

	// Ruta de health check para la base de datos
	r.GET("/health/db", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Verificar conexión a MongoDB
		dbStatus := "ok"
		dbError := ""

		if mongoClient == nil {
			dbStatus = "error"
			dbError = "MongoDB client not initialized"
		} else {
			// Intentar realizar una operación simple
			err := mongoClient.Ping(ctx, nil)
			if err != nil {
				dbStatus = "error"
				dbError = err.Error()
			}

			// Intentar listar las bases de datos
			_, err = mongoClient.ListDatabaseNames(ctx, bson.M{})
			if err != nil {
				dbStatus = "error"
				dbError = err.Error()
			}
		}

		// Construir respuesta
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
	})

	// Obtener puerto
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("📝 Using default port: %s", port)
	}

	// Inicializar repositorios
	userRepo := repository.NewUserRepository(mongoClient, os.Getenv("MONGODB_DATABASE"))

	// Inicializar handlers
	userHandler := handlers.NewUserHandler(userRepo)

	// Registrar rutas
	handlers.RegisterUserRoutes(r, userHandler)

	// Iniciar servidor
	serverAddr := ":" + port
	log.Printf("🚀 Server starting on http://localhost%s", serverAddr)
	log.Printf("💡 Try: curl http://localhost%s/health/db", serverAddr)

	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}
