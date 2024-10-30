// internal/handlers/user_handler.go
package handlers

import (
	"net/http"

	"github.com/ffelixf/microblog-platform/internal/models"
	"github.com/ffelixf/microblog-platform/internal/repository"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userRepo *repository.UserRepository
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userRepo.Create(c.Request.Context(), &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	user, err := h.userRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func RegisterUserRoutes(router *gin.Engine, handler *UserHandler) {
	api := router.Group("/api/v1")
	{
		api.POST("/users", handler.CreateUser)
		api.GET("/users/:id", handler.GetUser)
	}
}
