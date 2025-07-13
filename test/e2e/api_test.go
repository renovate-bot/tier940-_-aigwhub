package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"ai-gateway-hub/internal/config"
	"ai-gateway-hub/internal/database"
	"ai-gateway-hub/internal/handlers"
	"ai-gateway-hub/internal/middleware"
	"ai-gateway-hub/internal/services"
	"ai-gateway-hub/internal/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func setupTestServer(t *testing.T) (*gin.Engine, func()) {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "e2e_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Setup path manager
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	utils.InitPathManager()

	// Setup test configuration
	cfg := &config.Config{
		Port:                        "8080",
		SQLiteDBFile:               "./test.db",
		RedisAddr:                   "localhost:6379",
		LogDir:                      "./logs",
		LogLevel:                    "info",
		MaxSessions:                 100,
		SessionTimeout:              3600 * time.Second,
		WebSocketTimeout:            7200 * time.Second,
		ClaudeCLIPath:               "echo", // Use echo for testing
		GeminiCLIPath:               "echo",
		EnableProviderAutoDiscovery: true,
		EnableHealthChecks:          true,
	}

	// Initialize database
	db, err := database.InitSQLite(cfg.SQLiteDBFile)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	// Initialize Redis (will fail if not available, but test continues)
	redisClient := database.InitRedis(cfg.RedisAddr)

	// Initialize services
	sessionService := services.NewSessionService(redisClient)
	chatService := services.NewChatService(db)
	providerRegistry := services.NewProviderRegistry()

	// Register test providers
	if err := providerRegistry.RegisterDefaultProviders(cfg.LogDir); err != nil {
		t.Logf("Warning: Failed to register providers: %v", err)
	}

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.I18nMiddleware())
	router.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders: []string{"Content-Length"},
	}))

	// Setup routes
	router.GET("/", handlers.IndexHandler())
	router.GET("/chat/:id", handlers.ChatHandler(chatService))

	api := router.Group("/api")
	{
		api.GET("/health", handlers.HealthCheckHandler(redisClient))
		api.GET("/chats", handlers.GetChatsHandler(chatService))
		api.POST("/chats", handlers.CreateChatHandler(chatService))
		api.DELETE("/chats/:id", handlers.DeleteChatHandler(chatService))
		api.GET("/providers", handlers.GetProvidersHandler(providerRegistry))
	}

	// Initialize WebSocket hub
	hub := handlers.NewHub(sessionService, chatService, providerRegistry)
	go hub.Run()
	router.GET("/ws", handlers.WebSocketHandler(hub))

	// Cleanup function
	cleanup := func() {
		db.Close()
		redisClient.Close()
		os.Chdir(originalDir)
		os.RemoveAll(tempDir)
	}

	return router, cleanup
}

func TestHealthAPI(t *testing.T) {
	router, cleanup := setupTestServer(t)
	defer cleanup()

	t.Run("GET /api/health", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if status, ok := response["status"]; !ok || status != "ok" {
			t.Errorf("Expected status 'ok', got %v", status)
		}
	})
}

func TestProvidersAPI(t *testing.T) {
	router, cleanup := setupTestServer(t)
	defer cleanup()

	t.Run("GET /api/providers", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/providers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var providers []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &providers)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if len(providers) == 0 {
			t.Error("Expected at least one provider")
		}

		// Check that Claude provider is present
		found := false
		for _, provider := range providers {
			if id, ok := provider["id"]; ok && id == "claude" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Claude provider not found in response")
		}
	})
}

func TestChatsAPI(t *testing.T) {
	router, cleanup := setupTestServer(t)
	defer cleanup()

	var chatID float64

	t.Run("POST /api/chats - Create Chat", func(t *testing.T) {
		chatData := map[string]string{
			"title":    "Test Chat",
			"provider": "claude",
		}
		jsonData, _ := json.Marshal(chatData)

		req, _ := http.NewRequest("POST", "/api/chats", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if id, ok := response["id"]; ok {
			chatID = id.(float64)
		} else {
			t.Fatal("Response does not contain chat ID")
		}

		if title := response["title"]; title != "Test Chat" {
			t.Errorf("Expected title 'Test Chat', got %v", title)
		}

		if provider := response["provider"]; provider != "claude" {
			t.Errorf("Expected provider 'claude', got %v", provider)
		}
	})

	t.Run("GET /api/chats - List Chats", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/chats", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var chats []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &chats)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if len(chats) == 0 {
			t.Error("Expected at least one chat")
		}

		// Find our created chat
		found := false
		for _, chat := range chats {
			if id := chat["id"]; id == chatID {
				found = true
				if title := chat["title"]; title != "Test Chat" {
					t.Errorf("Expected title 'Test Chat', got %v", title)
				}
				break
			}
		}
		if !found {
			t.Error("Created chat not found in list")
		}
	})

	t.Run("GET /chat/:id - Chat Page", func(t *testing.T) {
		url := fmt.Sprintf("/chat/%d", int(chatID))
		req, _ := http.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Check that response contains HTML
		contentType := w.Header().Get("Content-Type")
		if contentType != "text/html; charset=utf-8" {
			t.Errorf("Expected HTML content type, got %s", contentType)
		}
	})

	t.Run("DELETE /api/chats/:id - Delete Chat", func(t *testing.T) {
		url := fmt.Sprintf("/api/chats/%d", int(chatID))
		req, _ := http.NewRequest("DELETE", url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if message := response["message"]; message != "Chat deleted successfully" {
			t.Errorf("Expected success message, got %v", message)
		}
	})

	t.Run("GET /chat/:id - Deleted Chat Returns 404", func(t *testing.T) {
		url := fmt.Sprintf("/chat/%d", int(chatID))
		req, _ := http.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404 for deleted chat, got %d", w.Code)
		}
	})
}

func TestCreateChatValidation(t *testing.T) {
	router, cleanup := setupTestServer(t)
	defer cleanup()

	t.Run("POST /api/chats - Missing Title", func(t *testing.T) {
		chatData := map[string]string{
			"provider": "claude",
		}
		jsonData, _ := json.Marshal(chatData)

		req, _ := http.NewRequest("POST", "/api/chats", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for missing title, got %d", w.Code)
		}
	})

	t.Run("POST /api/chats - Missing Provider", func(t *testing.T) {
		chatData := map[string]string{
			"title": "Test Chat",
		}
		jsonData, _ := json.Marshal(chatData)

		req, _ := http.NewRequest("POST", "/api/chats", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for missing provider, got %d", w.Code)
		}
	})

	t.Run("POST /api/chats - Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/chats", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid JSON, got %d", w.Code)
		}
	})
}

func TestIndexPage(t *testing.T) {
	router, cleanup := setupTestServer(t)
	defer cleanup()

	t.Run("GET / - Index Page", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Check that response contains HTML
		contentType := w.Header().Get("Content-Type")
		if contentType != "text/html; charset=utf-8" {
			t.Errorf("Expected HTML content type, got %s", contentType)
		}

		// Check that response contains expected content
		body := w.Body.String()
		if body == "" {
			t.Error("Response body is empty")
		}
	})
}

func TestCORSHeaders(t *testing.T) {
	router, cleanup := setupTestServer(t)
	defer cleanup()

	t.Run("OPTIONS /api/chats - CORS Preflight", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", "/api/chats", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status 204 for OPTIONS request, got %d", w.Code)
		}

		// Check CORS headers
		if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
			t.Errorf("Expected Access-Control-Allow-Origin '*', got '%s'", origin)
		}
	})
}