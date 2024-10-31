// internal/handlers/tweet_handler.go
package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ffelixf/microblog-platform/internal/models"
	"github.com/ffelixf/microblog-platform/internal/repository"
	"github.com/gin-gonic/gin"
)

type TweetHandler struct {
	tweetRepo *repository.TweetRepository
}

func NewTweetHandler(tweetRepo *repository.TweetRepository) *TweetHandler {
	return &TweetHandler{
		tweetRepo: tweetRepo,
	}
}

func (h *TweetHandler) CreateTweet(c *gin.Context) {
	var tweet models.Tweet
	if err := c.ShouldBindJSON(&tweet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.tweetRepo.Create(c.Request.Context(), &tweet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tweet)
}

func (h *TweetHandler) GetUserTweets(c *gin.Context) {
	userID := c.Param("id")

	tweets, err := h.tweetRepo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"count":   len(tweets),
		"tweets":  tweets,
	})
}

func (h *TweetHandler) GetTimeline(c *gin.Context) {
	userID := c.Param("id")

	// Obtener parámetros de paginación
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Validar parámetros
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	tweets, err := h.tweetRepo.GetTimeline(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Error al obtener timeline: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"page":    page,
		"limit":   limit,
		"count":   len(tweets),
		"tweets":  tweets,
	})
}

func RegisterTweetRoutes(router *gin.Engine, handler *TweetHandler) {
	api := router.Group("/api/v1")
	{
		api.POST("/tweets", handler.CreateTweet)
		api.GET("/users/:id/tweets", handler.GetUserTweets)
		api.GET("/users/:id/timeline", handler.GetTimeline)
	}
}
