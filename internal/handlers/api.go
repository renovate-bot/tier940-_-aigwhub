package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"ai-gateway-hub/internal/config"
	"ai-gateway-hub/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// HealthCheckHandler returns the health status
func HealthCheckHandler(redisClient *redis.Client, version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check Redis connection
		redisStatus := "healthy"
		if err := redisClient.Ping(c.Request.Context()).Err(); err != nil {
			redisStatus = "unhealthy: " + err.Error()
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"version": version,
			"redis":   redisStatus,
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

// GetProviderStatusHandler returns the cached status of a specific provider
func GetProviderStatusHandler(registry *services.ProviderRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		providerID := c.Param("id")
		
		provider, err := registry.Get(providerID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Provider not found",
			})
			return
		}
		
		// Use cached status for better performance
		status, err := registry.GetProviderStatus(providerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get provider status",
			})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"id":        provider.GetID(),
			"name":      provider.GetName(),
			"available": status.Available,
			"status":    status.Status,
			"version":   status.Version,
			"details":   status.Details,
		})
	}
}

// GetSettingsHandler returns current settings
func GetSettingsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get current language from context (set by i18n middleware)
		currentLang := c.GetString("lang")
		if currentLang == "" {
			currentLang = config.DefaultLanguage
		}
		
		// Get theme from cookie if available
		currentTheme := config.DefaultTheme
		if themeCookie, err := c.Cookie("theme"); err == nil && themeCookie != "" {
			currentTheme = themeCookie
		}
		
		c.JSON(http.StatusOK, gin.H{
			"language": currentLang,
			"theme":    currentTheme,
		})
	}
}

// UpdateSettingsHandler updates user settings
func UpdateSettingsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Language string `json:"language"`
			Theme    string `json:"theme"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request: " + err.Error(),
			})
			return
		}

		// Validate language
		if !config.IsValidLanguage(req.Language) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Unsupported language. Supported languages: " + strings.Join(config.SupportedLanguages, ", "),
			})
			return
		}

		// Validate theme
		if !config.IsValidTheme(req.Theme) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Unsupported theme. Supported themes: " + strings.Join(config.SupportedThemes, ", "),
			})
			return
		}

		// Set preference cookies with security flags
		secure := c.Request.TLS != nil // Use secure flag for HTTPS connections
		c.SetCookie("lang", req.Language, 30*24*3600, "/", "", secure, true)  // 30 days, httpOnly
		c.SetCookie("theme", req.Theme, 30*24*3600, "/", "", secure, true)    // 30 days, httpOnly
		
		c.JSON(http.StatusOK, gin.H{
			"message":  "Settings updated successfully",
			"language": req.Language,
			"theme":    req.Theme,
		})
	}
}