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
	"ai-gateway-hub/internal/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Version information (set during build)
var version = "dev"

//go:embed web/templates/*.html
var templateFiles embed.FS

//go:embed locales/*/*.json
var localeFiles embed.FS

//go:embed .env.example
var envExampleFile embed.FS

func main() {
	// Initialize path manager first
	if err := utils.InitPathManager(); err != nil {
		log.Fatalf("Failed to initialize path manager: %v", err)
	}

	// Load .env file if exists to get log configuration early
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or failed to load: %v", err)
	}

	// Load configuration
	cfg := config.Load()

	// Initialize logging early
	utils.InitLogger(cfg.LogLevel)
	
	// Initialize file logging
	if err := utils.InitFileLogging(cfg.LogDir); err != nil {
		log.Printf("Warning: Failed to initialize file logging: %v", err)
	} else {
		// Redirect standard log package to our custom logger
		utils.SetAsDefaultLogger()
	}
	
	utils.Info("AI Gateway Hub starting with log level: %s", cfg.LogLevel)

	// Initialize i18n first - extract files if needed and initialize once
	if err := initializeI18n(); err != nil {
		utils.Warn("Failed to initialize i18n: %v. Using default strings.", err)
	}

	// Extract .env.example (always update)
	if err := extractEnvExample(); err != nil {
		utils.Warn("Failed to extract .env.example: %v", err)
	}

	// Initialize database
	db, err := database.InitSQLite(cfg.SQLiteDBFile)
	if err != nil {
		utils.Fatal("Failed to initialize SQLite: %v", err)
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
	if err := providerRegistry.RegisterDefaultProviders(cfg); err != nil {
		utils.Warn("Failed to register default providers: %v", err)
	}

	// Setup logging level and Gin mode based on configuration
	setupLogging(cfg.LogLevel)

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
	
	// Create template with functions - language will be passed via template data
	tmpl := template.New("").Funcs(template.FuncMap{
		"T": func(lang interface{}, key string, args ...interface{}) string {
			langStr := "en"
			if lang != nil {
				if l, ok := lang.(string); ok && l != "" {
					langStr = l
				}
			}
			return i18n.T(langStr, key, args...)
		},
	})
	tmpl = template.Must(tmpl.ParseFS(templateFS, "*.html"))
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
		api.GET("/health", handlers.HealthCheckHandler(redisClient, version))
		api.GET("/chats", handlers.GetChatsHandler(chatService))
		api.POST("/chats", handlers.CreateChatHandler(chatService))
		api.DELETE("/chats/:id", handlers.DeleteChatHandler(chatService))
		api.GET("/providers", handlers.GetProvidersHandler(providerRegistry))
	}

	// WebSocket endpoint
	router.GET("/ws", handlers.WebSocketHandler(hub))

	// Get port from configuration
	port := cfg.Port

	// Create HTTP server with graceful shutdown support
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		utils.Info("Starting AI Gateway Hub on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Fatal("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	utils.Info("Shutting down server...")

	// Give the server 30 seconds to finish handling requests
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		utils.Fatal("Server forced to shutdown: %v", err)
	}

	utils.Info("Server exited")
}

// setupLogging configures Gin mode based on log level
func setupLogging(logLevel string) {
	switch logLevel {
	case "debug":
		utils.Debug("Setting Gin to debug mode")
		gin.SetMode(gin.DebugMode)
	case "info":
		utils.Debug("Setting Gin to release mode")
		gin.SetMode(gin.ReleaseMode)
	case "warn", "warning":
		utils.Debug("Setting Gin to release mode")
		gin.SetMode(gin.ReleaseMode)
	case "error":
		utils.Debug("Setting Gin to release mode")
		gin.SetMode(gin.ReleaseMode)
	default:
		utils.Warn("Unknown log level '%s', defaulting to INFO", logLevel)
		gin.SetMode(gin.ReleaseMode)
	}
}

// initializeI18n initializes i18n system with local files if they exist, otherwise embedded files
func initializeI18n() error {
	// Check if local locales directory exists and has files
	if _, err := os.Stat("locales/en/messages.json"); err == nil {
		// Use local files
		utils.Info("Using local i18n files from locales/ directory")
		return i18n.Init("locales", "en")
	} else {
		// Extract i18n files first, then use local files
		utils.Info("Extracting i18n files for customization")
		if err := extractI18nFiles(); err != nil {
			utils.Warn("Failed to extract i18n files, using embedded: %v", err)
			return i18n.InitWithFS(localeFiles, "en")
		}
		// Now use the extracted local files
		utils.Info("Using extracted i18n files from locales/ directory")
		return i18n.Init("locales", "en")
	}
}


// extractEnvExample extracts .env.example file (always overwrites)
func extractEnvExample() error {
	content, err := envExampleFile.ReadFile(".env.example")
	if err != nil {
		return err
	}

	if err := os.WriteFile(".env.example", content, 0644); err != nil {
		return err
	}

	utils.Info("Extracted .env.example for configuration reference")
	return nil
}

// extractI18nFiles extracts i18n files for user modification (only if they don't exist)
func extractI18nFiles() error {
	// Create locales directory if it doesn't exist
	if err := utils.EnsureDir("locales"); err != nil {
		return err
	}

	// Walk through embedded locale files
	return fs.WalkDir(localeFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// Create directory
			if err := utils.EnsureDir(path); err != nil {
				return err
			}
			return nil
		}

		// Check if file already exists (don't overwrite user modifications)
		localPath := path
		if _, err := os.Stat(localPath); err == nil {
			// File exists, skip extraction
			return nil
		}

		// Extract file
		content, err := localeFiles.ReadFile(path)
		if err != nil {
			return err
		}

		if err := utils.WriteToFile(localPath, content); err != nil {
			return err
		}

		utils.Info("Extracted i18n file: %s", localPath)
		return nil
	})
}