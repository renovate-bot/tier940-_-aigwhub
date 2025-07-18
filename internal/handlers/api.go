package handlers

import (
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"ai-gateway-hub/internal/config"
	"ai-gateway-hub/internal/services"
	"ai-gateway-hub/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const (
	// Cookie expiration time (30 days)
	CookieMaxAge = 30 * 24 * 3600
)

// APIHandlers contains the dependencies for all API handlers
type APIHandlers struct {
	errorHandler *ErrorHandler
}

// NewAPIHandlers creates a new APIHandlers instance with proper dependencies
func NewAPIHandlers(logger *log.Logger) *APIHandlers {
	return &APIHandlers{
		errorHandler: NewErrorHandler(logger),
	}
}

// HealthCheckHandler returns the health status
func HealthCheckHandler(redisClient *redis.Client, version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check Redis connection
		redisStatus := "healthy"
		if err := redisClient.Ping(c.Request.Context()).Err(); err != nil {
			redisStatus = "unhealthy"
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"version": version,
			"redis":   redisStatus,
		})
	}
}

// GetChatsHandler returns list of chats
func (h *APIHandlers) GetChatsHandler(chatService *services.ChatService) gin.HandlerFunc {
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
			h.errorHandler.InternalError(c, "Failed to get chats", err)
			return
		}

		h.errorHandler.Success(c, chats)
	}
}

// CreateChatHandler creates a new chat
func (h *APIHandlers) CreateChatHandler(chatService *services.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Title    string `json:"title" binding:"required"`
			Provider string `json:"provider" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			h.errorHandler.ValidationError(c, "Invalid request", err)
			return
		}

		chat, err := chatService.CreateChat(req.Title, req.Provider)
		if err != nil {
			h.errorHandler.InternalError(c, "Failed to create chat", err)
			return
		}

		h.errorHandler.Created(c, chat, "Chat created successfully")
	}
}

// DeleteChatHandler deletes a chat
func (h *APIHandlers) DeleteChatHandler(chatService *services.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatIDStr := c.Param("id")
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			h.errorHandler.BadRequest(c, "Invalid chat ID", err)
			return
		}

		if err := chatService.DeleteChat(chatID); err != nil {
			h.errorHandler.InternalError(c, "Failed to delete chat", err)
			return
		}

		h.errorHandler.Success(c, nil, "Chat deleted successfully")
	}
}

// GetProvidersHandler returns available AI providers
func (h *APIHandlers) GetProvidersHandler(registry *services.ProviderRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		providers := registry.List()
		h.errorHandler.Success(c, providers)
	}
}

// GetProviderStatusHandler returns the cached status of a specific provider
func (h *APIHandlers) GetProviderStatusHandler(registry *services.ProviderRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		providerID := c.Param("id")
		
		provider, err := registry.Get(providerID)
		if err != nil {
			h.errorHandler.NotFound(c, "Provider not found")
			return
		}
		
		// Use cached status for better performance
		status, err := registry.GetProviderStatus(providerID)
		if err != nil {
			h.errorHandler.InternalError(c, "Failed to get provider status", err)
			return
		}
		
		response := gin.H{
			"id":        provider.GetID(),
			"name":      provider.GetName(),
			"available": status.Available,
			"status":    status.Status,
			"version":   status.Version,
			"details":   status.Details,
		}
		h.errorHandler.Success(c, response)
	}
}

// GetSettingsHandler returns current settings
func (h *APIHandlers) GetSettingsHandler() gin.HandlerFunc {
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
		
		// Get chat input behavior from cookie if available
		currentChatBehavior := "enter_to_send" // Default
		if chatBehaviorCookie, err := c.Cookie("chatInputBehavior"); err == nil && chatBehaviorCookie != "" {
			currentChatBehavior = chatBehaviorCookie
		}
		
		settings := gin.H{
			"language": currentLang,
			"theme":    currentTheme,
			"chatInputBehavior": currentChatBehavior,
		}
		h.errorHandler.Success(c, settings)
	}
}

// UpdateSettingsHandler updates user settings
func (h *APIHandlers) UpdateSettingsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Language          string `json:"language"`
			Theme            string `json:"theme"`
			ChatInputBehavior string `json:"chatInputBehavior"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			h.errorHandler.ValidationError(c, "Invalid request", err)
			return
		}

		// Set defaults for empty fields
		if req.Language == "" {
			req.Language = config.DefaultLanguage
		}
		if req.Theme == "" {
			req.Theme = config.DefaultTheme
		}
		if req.ChatInputBehavior == "" {
			req.ChatInputBehavior = "enter_to_send"
		}

		// Validate language
		if !config.IsValidLanguage(req.Language) {
			h.errorHandler.BadRequest(c, "Unsupported language. Supported languages: "+strings.Join(config.SupportedLanguages, ", "), nil)
			return
		}

		// Validate theme
		if !config.IsValidTheme(req.Theme) {
			h.errorHandler.BadRequest(c, "Unsupported theme. Supported themes: "+strings.Join(config.SupportedThemes, ", "), nil)
			return
		}

		// Validate chat input behavior
		validInputBehaviors := []string{"enter_to_send", "ctrl_enter_to_send"}
		if req.ChatInputBehavior != "" {
			if !slices.Contains(validInputBehaviors, req.ChatInputBehavior) {
				h.errorHandler.BadRequest(c, "Invalid chat input behavior. Supported: "+strings.Join(validInputBehaviors, ", "), nil)
				return
			}
		}

		// Set preference cookies with security flags
		secure := c.Request.TLS != nil // Use secure flag for HTTPS connections
		c.SetCookie("lang", req.Language, CookieMaxAge, "/", "", secure, true)  // 30 days, httpOnly
		c.SetCookie("theme", req.Theme, CookieMaxAge, "/", "", secure, true)    // 30 days, httpOnly
		c.SetCookie("chatInputBehavior", req.ChatInputBehavior, CookieMaxAge, "/", "", secure, true) // 30 days, httpOnly
		
		response := gin.H{
			"language": req.Language,
			"theme":    req.Theme,
			"chatInputBehavior": req.ChatInputBehavior,
		}
		h.errorHandler.Success(c, response, "Settings updated successfully")
	}
}

// LogClientErrorHandler logs client-side errors to server logs
func (h *APIHandlers) LogClientErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Message string `json:"message"`
			Stack   string `json:"stack"`
			URL     string `json:"url"`
			UserAgent string `json:"userAgent"`
			Level   string `json:"level"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			h.errorHandler.BadRequest(c, "Invalid request", err)
			return
		}

		// Log the client error to server logs
		clientInfo := fmt.Sprintf("URL: %s, User-Agent: %s", req.URL, req.UserAgent)
		errorMessage := fmt.Sprintf("Client Error: %s", req.Message)
		if req.Stack != "" {
			errorMessage += fmt.Sprintf("\nStack: %s", req.Stack)
		}
		errorMessage += fmt.Sprintf("\n%s", clientInfo)

		switch req.Level {
		case "error":
			utils.Error(errorMessage)
		case "warn":
			utils.Warn(errorMessage)
		case "info":
			utils.Info(errorMessage)
		default:
			utils.Error(errorMessage)
		}

		h.errorHandler.Success(c, nil, "Error logged")
	}
}