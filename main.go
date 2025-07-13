package main

import (
	"context"
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-gateway-hub/internal/config"
	"ai-gateway-hub/internal/database"
	"ai-gateway-hub/internal/handlers"
	"ai-gateway-hub/internal/i18n"
	"ai-gateway-hub/internal/middleware"
	"ai-gateway-hub/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

//go:embed web/templates/*.html
var templateFiles embed.FS

//go:embed locales/*/*.json
var localeFiles embed.FS

func main() {
	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or failed to load: %v", err)
	}

	// Load configuration
	cfg := config.Load()

	// Initialize i18n with embedded files
	if err := i18n.InitWithFS(localeFiles, "en"); err != nil {
		log.Printf("Warning: Failed to initialize i18n: %v. Using default strings.", err)
	}

	// Initialize database
	db, err := database.InitSQLite(cfg.SQLiteDBPath)
	if err != nil {
		log.Fatalf("Failed to initialize SQLite: %v", err)
	}
	defer db.Close()

	// Initialize Redis
	redisClient := database.InitRedis(cfg.RedisAddr)
	defer redisClient.Close()

	// Initialize services
	sessionService := services.NewSessionService(redisClient)
	chatService := services.NewChatService(db)
	providerRegistry := services.NewProviderRegistry()
	
	// Register providers
	if err := providerRegistry.RegisterDefaultProviders(); err != nil {
		log.Printf("Warning: Failed to register default providers: %v", err)
	}

	// Initialize Gin router
	router := gin.Default()

	// Setup middleware
	router.Use(middleware.I18nMiddleware())

	// Setup CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Static files are served from external directory if needed
	// For now, static files are handled by external CDN (Tailwind CSS, Alpine.js)

	// Load embedded HTML templates
	templateFS, err := fs.Sub(templateFiles, "web/templates")
	if err != nil {
		log.Fatalf("Failed to create template file system: %v", err)
	}
	
	tmpl := template.Must(template.New("").ParseFS(templateFS, "*.html"))
	router.SetHTMLTemplate(tmpl)

	// Initialize WebSocket hub
	hub := handlers.NewHub(sessionService, chatService, providerRegistry)
	go hub.Run()

	// Setup routes
	router.GET("/", handlers.IndexHandler())
	router.GET("/chat/:id", handlers.ChatHandler(chatService))

	// API routes
	api := router.Group("/api")
	{
		api.GET("/health", handlers.HealthCheckHandler(redisClient))
		api.GET("/chats", handlers.GetChatsHandler(chatService))
		api.POST("/chats", handlers.CreateChatHandler(chatService))
		api.DELETE("/chats/:id", handlers.DeleteChatHandler(chatService))
		api.GET("/providers", handlers.GetProvidersHandler(providerRegistry))
	}

	// WebSocket endpoint
	router.GET("/ws", handlers.WebSocketHandler(hub))

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server with graceful shutdown support
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting AI Gateway Hub on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give the server 30 seconds to finish handling requests
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}