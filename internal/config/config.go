package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server settings
	Port string

	// Database paths
	SQLiteDBPath string
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

	// Feature flags
	EnableProviderAutoDiscovery bool
	EnableHealthChecks          bool
}

func Load() *Config {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		SQLiteDBPath: getEnv("SQLITE_DB_PATH", "./data/ai_gateway.db"),
		RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
		StaticDir:    getEnv("STATIC_DIR", "./web/static"),
		TemplateDir:  getEnv("TEMPLATE_DIR", "./web/templates"),
		LogDir:       getEnv("LOG_DIR", "./logs"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),

		MaxSessions:      getEnvAsInt("MAX_SESSIONS", 100),
		SessionTimeout:   time.Duration(getEnvAsInt("SESSION_TIMEOUT", 3600)) * time.Second,
		WebSocketTimeout: time.Duration(getEnvAsInt("WEBSOCKET_TIMEOUT", 7200)) * time.Second,

		ClaudeCLIPath: getEnv("CLAUDE_CLI_PATH", "claude"),
		GeminiCLIPath: getEnv("GEMINI_CLI_PATH", "gemini"),

		EnableProviderAutoDiscovery: getEnvAsBool("ENABLE_PROVIDER_AUTO_DISCOVERY", true),
		EnableHealthChecks:          getEnvAsBool("ENABLE_HEALTH_CHECKS", true),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}