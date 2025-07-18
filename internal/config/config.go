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
	// Create new instance to avoid global state issues in tests
	v := viper.New()
	
	// Set configuration name and type
	v.SetConfigName(".env")
	v.SetConfigType("env")
	
	// Add config path
	v.AddConfigPath(".")
	
	// Set default values
	setDefaultsForViper(v)
	
	// Enable environment variable reading
	v.AutomaticEnv()
	
	// Read configuration file if it exists
	if err := v.ReadInConfig(); err != nil {
		// Config file not found or error reading - use defaults and env vars
	}
	
	// Helper function to get int with fallback to default
	getIntWithDefault := func(key string, defaultValue int) int {
		val := v.GetInt(key)
		if val == 0 && v.GetString(key) != "0" && v.GetString(key) != "" {
			// Value is invalid, return default
			return defaultValue
		}
		return val
	}
	
	// Helper function to get bool with fallback to default
	getBoolWithDefault := func(key string, defaultValue bool) bool {
		str := v.GetString(key)
		if str == "" {
			return defaultValue
		}
		if str == "true" || str == "1" {
			return true
		}
		if str == "false" || str == "0" {
			return false
		}
		// Invalid value, return default
		return defaultValue
	}
	
	return &Config{
		Port:         v.GetString("PORT"),
		SQLiteDBFile: v.GetString("SQLITE_DB_FILE"),
		RedisAddr:    v.GetString("REDIS_ADDR"),
		StaticDir:    v.GetString("STATIC_DIR"),
		TemplateDir:  v.GetString("TEMPLATE_DIR"),
		LogDir:       v.GetString("LOG_DIR"),
		LogLevel:     v.GetString("LOG_LEVEL"),

		MaxSessions:      getIntWithDefault("MAX_SESSIONS", 100),
		SessionTimeout:   time.Duration(getIntWithDefault("SESSION_TIMEOUT", 3600)) * time.Second,
		WebSocketTimeout: time.Duration(getIntWithDefault("WEBSOCKET_TIMEOUT", 7200)) * time.Second,

		ClaudeCLIPath: v.GetString("CLAUDE_CLI_PATH"),
		GeminiCLIPath: v.GetString("GEMINI_CLI_PATH"),

		ClaudeSkipPermissions: getBoolWithDefault("CLAUDE_SKIP_PERMISSIONS", false),
		ClaudeExtraArgs:       v.GetString("CLAUDE_EXTRA_ARGS"),

		EnableProviderAutoDiscovery: getBoolWithDefault("ENABLE_PROVIDER_AUTO_DISCOVERY", true),
		EnableHealthChecks:          getBoolWithDefault("ENABLE_HEALTH_CHECKS", true),
	}
}

// setDefaults sets default configuration values
func setDefaults() {
	setDefaultsForViper(viper.GetViper())
}

// setDefaultsForViper sets default configuration values for a specific viper instance
func setDefaultsForViper(v *viper.Viper) {
	// Server Configuration
	v.SetDefault("PORT", "8080")
	v.SetDefault("SQLITE_DB_FILE", "./data/ai_gateway.db")
	v.SetDefault("REDIS_ADDR", "localhost:6379")
	v.SetDefault("STATIC_DIR", "./web/static")
	v.SetDefault("TEMPLATE_DIR", "./web/templates")
	
	// Logging Configuration
	v.SetDefault("LOG_DIR", "./logs")
	v.SetDefault("LOG_LEVEL", "info")
	
	// Session Management
	v.SetDefault("MAX_SESSIONS", 100)
	v.SetDefault("SESSION_TIMEOUT", 3600)
	v.SetDefault("WEBSOCKET_TIMEOUT", 7200)
	
	// AI Provider Configuration
	v.SetDefault("CLAUDE_CLI_PATH", "claude")
	v.SetDefault("GEMINI_CLI_PATH", "gemini")
	
	// Claude CLI Options
	v.SetDefault("CLAUDE_SKIP_PERMISSIONS", false)
	v.SetDefault("CLAUDE_EXTRA_ARGS", "")
	
	// Feature Flags
	v.SetDefault("ENABLE_PROVIDER_AUTO_DISCOVERY", true)
	v.SetDefault("ENABLE_HEALTH_CHECKS", true)
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

