// cmd/api/main.go
package main

import (
    "context"
    "log"
    "os"
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    "github.com/ffelixf/microblog-platform/pkg/database"
    "github.com/ffelixf/microblog-platform/internal/handlers"
)

func main() {
    // Cargar variables de entorno
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    // Inicializar router
    r := gin.Default()

    // Conectar a MongoDB
    mongoClient, err := database.ConnectMongoDB()
    if err != nil {
        log.Fatal("Failed to connect to MongoDB:", err)
    }
    defer mongoClient.Disconnect(context.Background())

    // Configurar rutas
    setupRoutes(r)

    // Iniciar servidor
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    log.Printf("Server starting on port %s", port)
    if err := r.Run(":" + port); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}

func setupRoutes(r *gin.Engine) {
    // Ruta de health check
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "ok",
        })
    })

    // Grupo de rutas API v1
    v1 := r.Group("/api/v1")
    {
        // Las rutas se agregarán aquí
    }
}