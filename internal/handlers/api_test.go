package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ai-gateway-hub/internal/database"
	"ai-gateway-hub/internal/models"
	"ai-gateway-hub/internal/providers"
	"ai-gateway-hub/internal/services"
)

func setupAPITest(t *testing.T) (*gin.Engine, *services.ChatService, func()) {
	gin.SetMode(gin.TestMode)

	db, err := database.InitTestDB()
	require.NoError(t, err)

	chatService := services.NewChatService(db)
	registry := providers.NewRegistry()

	router := gin.New()
	
	// Register API endpoints
	api := router.Group("/api")
	{
		api.GET("/chats", listChatsHandler(chatService))
		api.POST("/chats", createChatHandler(chatService))
		api.DELETE("/chats/:id", deleteChatHandler(chatService))
		api.GET("/providers", listProvidersHandler(registry))
		api.GET("/health", healthCheckHandler())
	}

	cleanup := func() {
		db.Close()
	}

	return router, chatService, cleanup
}

// Handler functions for testing
func listChatsHandler(chatService *services.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := 10
		offset := 0
		
		if l := c.Query("limit"); l != "" {
			if val, err := strconv.Atoi(l); err == nil {
				if val <= 0 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
					return
				}
				limit = val
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
				return
			}
		}
		
		if o := c.Query("offset"); o != "" {
			if val, err := strconv.Atoi(o); err == nil {
				if val < 0 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset"})
					return
				}
				offset = val
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset"})
				return
			}
		}
		
		chats, err := chatService.GetChats(limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, chats)
	}
}

func createChatHandler(chatService *services.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Content-Type") != "application/json" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Content-Type must be application/json"})
			return
		}
		
		var req struct {
			Title    string `json:"title"`
			Provider string `json:"provider"`
		}
		
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}
		
		if req.Provider == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Provider is required"})
			return
		}
		
		chat, err := chatService.CreateChat(req.Title, req.Provider)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusCreated, chat)
	}
}

func deleteChatHandler(chatService *services.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
			return
		}
		
		err = chatService.DeleteChat(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.Status(http.StatusNoContent)
	}
}

func listProvidersHandler(registry *providers.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		providerList := registry.List()
		providers := make([]map[string]interface{}, 0, len(providerList))
		
		for _, p := range providerList {
			providers = append(providers, map[string]interface{}{
				"name":      p.GetName(),
				"available": p.IsAvailable(),
			})
		}
		
		c.JSON(http.StatusOK, gin.H{"providers": providers})
	}
}

func healthCheckHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": http.TimeFormat,
		})
	}
}

func TestListChats(t *testing.T) {
	router, chatService, cleanup := setupAPITest(t)
	defer cleanup()

	// Create test chats
	for i := 0; i < 5; i++ {
		_, err := chatService.CreateChat("Test Chat "+string(rune('A'+i)), "claude")
		require.NoError(t, err)
	}

	tests := []struct {
		name           string
		query          string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "list all chats default",
			query:          "",
			expectedStatus: http.StatusOK,
			expectedCount:  5,
		},
		{
			name:           "list chats with limit",
			query:          "?limit=3",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "list chats with offset",
			query:          "?offset=3",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "list chats with limit and offset",
			query:          "?limit=2&offset=2",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "list chats with invalid limit",
			query:          "?limit=abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "list chats with negative limit",
			query:          "?limit=-1",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/chats"+tt.query, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedStatus, resp.Code)

			if tt.expectedStatus == http.StatusOK {
				var chats []models.Chat
				err := json.Unmarshal(resp.Body.Bytes(), &chats)
				assert.NoError(t, err)
				assert.Len(t, chats, tt.expectedCount)
			}
		})
	}
}

func TestCreateChat(t *testing.T) {
	router, _, cleanup := setupAPITest(t)
	defer cleanup()

	tests := []struct {
		name           string
		payload        interface{}
		contentType    string
		expectedStatus int
		validate       func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "create valid chat",
			payload: map[string]interface{}{
				"title":    "New Chat",
				"provider": "claude",
			},
			contentType:    "application/json",
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var chat models.Chat
				err := json.Unmarshal(resp.Body.Bytes(), &chat)
				assert.NoError(t, err)
				assert.NotZero(t, chat.ID)
				assert.Equal(t, "New Chat", chat.Title)
				assert.Equal(t, "claude", chat.Provider)
			},
		},
		{
			name: "create chat with empty title",
			payload: map[string]interface{}{
				"title":    "",
				"provider": "gemini",
			},
			contentType:    "application/json",
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var chat models.Chat
				err := json.Unmarshal(resp.Body.Bytes(), &chat)
				assert.NoError(t, err)
				assert.Equal(t, "", chat.Title)
			},
		},
		{
			name: "create chat without provider",
			payload: map[string]interface{}{
				"title": "No Provider",
			},
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "create chat with invalid JSON",
			payload:        "invalid json",
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "create chat without content-type",
			payload: map[string]interface{}{
				"title":    "Test",
				"provider": "claude",
			},
			contentType:    "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if str, ok := tt.payload.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.payload)
			}

			req, _ := http.NewRequest("POST", "/api/chats", bytes.NewBuffer(body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedStatus, resp.Code)

			if tt.validate != nil {
				tt.validate(t, resp)
			}
		})
	}
}

func TestDeleteChat(t *testing.T) {
	router, chatService, cleanup := setupAPITest(t)
	defer cleanup()

	// Create test chats
	chat1, err := chatService.CreateChat("Chat to Delete", "claude")
	require.NoError(t, err)

	chat2, err := chatService.CreateChat("Chat to Keep", "gemini")
	require.NoError(t, err)

	tests := []struct {
		name           string
		chatID         string
		expectedStatus int
		verify         func(t *testing.T)
	}{
		{
			name:           "delete existing chat",
			chatID:         strconv.FormatInt(chat1.ID, 10),
			expectedStatus: http.StatusNoContent,
			verify: func(t *testing.T) {
				// Verify chat is deleted
				_, err := chatService.GetChat(chat1.ID)
				assert.Error(t, err)

				// Verify other chat still exists
				_, err = chatService.GetChat(chat2.ID)
				assert.NoError(t, err)
			},
		},
		{
			name:           "delete non-existing chat",
			chatID:         "99999",
			expectedStatus: http.StatusNoContent, // Idempotent
		},
		{
			name:           "delete with invalid ID",
			chatID:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "delete with negative ID",
			chatID:         "-1",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("DELETE", "/api/chats/"+tt.chatID, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedStatus, resp.Code)

			if tt.verify != nil {
				tt.verify(t)
			}
		})
	}
}

func TestListProviders(t *testing.T) {
	router, _, cleanup := setupAPITest(t)
	defer cleanup()

	// Register some mock providers
	registry := providers.NewRegistry()
	registry.Register("claude", &mockAIProvider{name: "claude", healthy: true})
	registry.Register("gemini", &mockAIProvider{name: "gemini", healthy: false})

	// Update router with new registry
	router = gin.New()
	router.GET("/api/providers", listProvidersHandler(registry))

	req, _ := http.NewRequest("GET", "/api/providers", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)

	providers, ok := response["providers"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, providers, 2)
}

func TestHealthCheck(t *testing.T) {
	router, _, cleanup := setupAPITest(t)
	defer cleanup()

	req, _ := http.NewRequest("GET", "/api/health", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
	assert.NotEmpty(t, response["timestamp"])
}

// Helper mock provider for testing
type mockAIProvider struct {
	name    string
	healthy bool
}

func (m *mockAIProvider) GetID() string {
	return m.name
}

func (m *mockAIProvider) GetName() string {
	return m.name
}

func (m *mockAIProvider) GetDescription() string {
	return "Mock AI Provider for testing"
}

func (m *mockAIProvider) IsAvailable() bool {
	return m.healthy
}

func (m *mockAIProvider) GetStatus() providers.ProviderStatus {
	status := "ready"
	if !m.healthy {
		status = "error"
	}
	return providers.ProviderStatus{
		Available: m.healthy,
		Status:    status,
		Version:   "1.0.0",
		Details:   "Mock provider",
	}
}

func (m *mockAIProvider) SendPrompt(ctx context.Context, prompt string, chatID int64) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("Mock response")), nil
}

func (m *mockAIProvider) StreamResponse(ctx context.Context, prompt string, chatID int64, writer io.Writer) error {
	_, err := writer.Write([]byte("Mock streaming response"))
	return err
}