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

// CreateUser godoc
// @Summary      Crear nuevo usuario
// @Description  Crea un nuevo usuario en la plataforma
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      models.User  true  "Información del usuario"
// @Success      201   {object}  models.User
// @Failure      400   {object}  models.Error
// @Failure      409   {object}  models.Error
// @Router       /users [post]

// CreateUser maneja la creación de nuevos usuarios
func (h *UserHandler) CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := h.userRepo.Create(c.Request.Context(), &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error al crear usuario: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// GetUser godoc
// @Summary      Obtener usuario por ID
// @Description  Obtiene un usuario por su ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "ID del usuario"
// @Success      200  {object}  models.User
// @Failure      404  {object}  models.Error
// @Router       /users/{id} [get]

// GetUser maneja la obtención de un usuario por ID
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	user, err := h.userRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Usuario no encontrado",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetFollowing obtiene la lista de usuarios que sigue un usuario
func (h *UserHandler) GetFollowing(c *gin.Context) {
	userID := c.Param("id")

	following, err := h.userRepo.GetFollowing(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Siempre retornar una respuesta estructurada
	c.JSON(http.StatusOK, gin.H{
		"user_id":   userID,
		"count":     len(following),
		"following": following,
	})
}

// GetFollowers obtiene la lista de seguidores de un usuario
func (h *UserHandler) GetFollowers(c *gin.Context) {
	userID := c.Param("id")

	followers, err := h.userRepo.GetFollowers(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error al obtener seguidores: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":     len(followers),
		"followers": followers,
	})
}

// FollowUser godoc
// @Summary      Seguir a un usuario
// @Description  Hace que un usuario siga a otro
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id         path      string  true  "ID del usuario que sigue"
// @Param        target_id  path      string  true  "ID del usuario a seguir"
// @Success      200        {object}  models.FollowResponse
// @Failure      400        {object}  models.Error
// @Failure      404        {object}  models.Error
// @Router       /users/{id}/follow/{target_id} [post]

// FollowUser maneja la acción de seguir a otro usuario
func (h *UserHandler) FollowUser(c *gin.Context) {
	userID := c.Param("id")
	targetID := c.Param("target_id")

	if err := h.userRepo.FollowUser(c.Request.Context(), userID, targetID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Error al seguir usuario: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Usuario seguido exitosamente",
		"user_id":      userID,
		"following_id": targetID,
	})
}

// UnfollowUser maneja la acción de dejar de seguir a otro usuario
func (h *UserHandler) UnfollowUser(c *gin.Context) {
	userID := c.Param("id")
	targetID := c.Param("target_id")

	if err := h.userRepo.UnfollowUser(c.Request.Context(), userID, targetID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Error al dejar de seguir usuario: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Usuario dejado de seguir exitosamente",
		"user_id":       userID,
		"unfollowed_id": targetID,
	})
}

// RegisterUserRoutes registra todas las rutas relacionadas con usuarios
func RegisterUserRoutes(router *gin.Engine, handler *UserHandler) {
	api := router.Group("/api/v1")
	{
		// Rutas básicas de usuarios
		api.POST("/users", handler.CreateUser)
		api.GET("/users/:id", handler.GetUser)

		// Rutas de following/followers
		api.POST("/users/:id/follow/:target_id", handler.FollowUser)
		api.POST("/users/:id/unfollow/:target_id", handler.UnfollowUser)
		api.GET("/users/:id/following", handler.GetFollowing)
		api.GET("/users/:id/followers", handler.GetFollowers)
	}
}
