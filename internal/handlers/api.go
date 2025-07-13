package handlers

import (
	"net/http"
	"strconv"

	"ai-gateway-hub/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// HealthCheckHandler returns the health status
func HealthCheckHandler(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check Redis connection
		redisStatus := "healthy"
		if err := redisClient.Ping(c.Request.Context()).Err(); err != nil {
			redisStatus = "unhealthy: " + err.Error()
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"redis":  redisStatus,
		})
	}
}

// GetChatsHandler returns list of chats
func GetChatsHandler(chatService *services.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := 50
		offset := 0

		if l := c.Query("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
				limit = parsed
			}
		}

		if o := c.Query("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
				offset = parsed
			}
		}

		chats, err := chatService.GetChats(limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get chats",
			})
			return
		}

		c.JSON(http.StatusOK, chats)
	}
}

// CreateChatHandler creates a new chat
func CreateChatHandler(chatService *services.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Title    string `json:"title" binding:"required"`
			Provider string `json:"provider" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request: " + err.Error(),
			})
			return
		}

		chat, err := chatService.CreateChat(req.Title, req.Provider)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create chat",
			})
			return
		}

		c.JSON(http.StatusCreated, chat)
	}
}

// DeleteChatHandler deletes a chat
func DeleteChatHandler(chatService *services.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatIDStr := c.Param("id")
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid chat ID",
			})
			return
		}

		if err := chatService.DeleteChat(chatID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete chat",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Chat deleted successfully",
		})
	}
}

// GetProvidersHandler returns available AI providers
func GetProvidersHandler(registry *services.ProviderRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		providers := registry.List()
		c.JSON(http.StatusOK, providers)
	}
}