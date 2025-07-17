package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	// Server settings
	Port string

	// Database settings
	SQLiteDBFile string
	RedisAddr    string

	// Static files
	StaticDir   string
	TemplateDir string

	// Log settings
	LogDir   string
	LogLevel string

	// Session management
	MaxSessions      int
	SessionTimeout   time.Duration
	WebSocketTimeout time.Duration

	// AI Provider paths
	ClaudeCLIPath string
	GeminiCLIPath string

	// Claude CLI Options
	ClaudeSkipPermissions bool
	ClaudeExtraArgs       string

	// Feature flags
	EnableProviderAutoDiscovery bool
	EnableHealthChecks          bool
}

// Load initializes and loads configuration from various sources
func Load() *Config {
	// Set configuration name and type
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	
	// Add config path
	viper.AddConfigPath(".")
	
	// Set default values
	setDefaults()
	
	// Enable environment variable reading
	viper.AutomaticEnv()
	
	// Read configuration file if it exists
	if err := viper.ReadInConfig(); err != nil {
		// Config file not found or error reading - use defaults and env vars
	}
	
	return &Config{
		Port:         viper.GetString("PORT"),
		SQLiteDBFile: viper.GetString("SQLITE_DB_FILE"),
		RedisAddr:    viper.GetString("REDIS_ADDR"),
		StaticDir:    viper.GetString("STATIC_DIR"),
		TemplateDir:  viper.GetString("TEMPLATE_DIR"),
		LogDir:       viper.GetString("LOG_DIR"),
		LogLevel:     viper.GetString("LOG_LEVEL"),

		MaxSessions:      viper.GetInt("MAX_SESSIONS"),
		SessionTimeout:   time.Duration(viper.GetInt("SESSION_TIMEOUT")) * time.Second,
		WebSocketTimeout: time.Duration(viper.GetInt("WEBSOCKET_TIMEOUT")) * time.Second,

		ClaudeCLIPath: viper.GetString("CLAUDE_CLI_PATH"),
		GeminiCLIPath: viper.GetString("GEMINI_CLI_PATH"),

		ClaudeSkipPermissions: viper.GetBool("CLAUDE_SKIP_PERMISSIONS"),
		ClaudeExtraArgs:       viper.GetString("CLAUDE_EXTRA_ARGS"),

		EnableProviderAutoDiscovery: viper.GetBool("ENABLE_PROVIDER_AUTO_DISCOVERY"),
		EnableHealthChecks:          viper.GetBool("ENABLE_HEALTH_CHECKS"),
	}
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server Configuration
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("SQLITE_DB_FILE", "./data/ai_gateway.db")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("STATIC_DIR", "./web/static")
	viper.SetDefault("TEMPLATE_DIR", "./web/templates")
	
	// Logging Configuration
	viper.SetDefault("LOG_DIR", "./logs")
	viper.SetDefault("LOG_LEVEL", "info")
	
	// Session Management
	viper.SetDefault("MAX_SESSIONS", 100)
	viper.SetDefault("SESSION_TIMEOUT", 3600)
	viper.SetDefault("WEBSOCKET_TIMEOUT", 7200)
	
	// AI Provider Configuration
	viper.SetDefault("CLAUDE_CLI_PATH", "claude")
	viper.SetDefault("GEMINI_CLI_PATH", "gemini")
	
	// Claude CLI Options
	viper.SetDefault("CLAUDE_SKIP_PERMISSIONS", false)
	viper.SetDefault("CLAUDE_EXTRA_ARGS", "")
	
	// Feature Flags
	viper.SetDefault("ENABLE_PROVIDER_AUTO_DISCOVERY", true)
	viper.SetDefault("ENABLE_HEALTH_CHECKS", true)
}

// GetString returns a configuration value as string with environment variable support
func GetString(key string) string {
	return viper.GetString(key)
}

// GetInt returns a configuration value as int with environment variable support
func GetInt(key string) int {
	return viper.GetInt(key)
}

// GetBool returns a configuration value as bool with environment variable support
func GetBool(key string) bool {
	return viper.GetBool(key)
}

// SetConfigPath adds an additional path to search for config files
func SetConfigPath(path string) {
	viper.AddConfigPath(path)
}

// IsProduction returns true if the application is running in production mode
func IsProduction() bool {
	env := strings.ToLower(viper.GetString("GIN_MODE"))
	return env == "release" || env == "production"
}

// IsDevelopment returns true if the application is running in development mode
func IsDevelopment() bool {
	return !IsProduction()
}

