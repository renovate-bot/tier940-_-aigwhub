package unit

import (
	"os"
	"testing"
	"time"

	"ai-gateway-hub/internal/config"
)

func TestConfigLoad(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"PORT", "SQLITE_DB_FILE", "REDIS_ADDR", "STATIC_DIR", "TEMPLATE_DIR",
		"LOG_DIR", "LOG_LEVEL", "MAX_SESSIONS", "SESSION_TIMEOUT", "WEBSOCKET_TIMEOUT",
		"CLAUDE_CLI_PATH", "GEMINI_CLI_PATH", "ENABLE_PROVIDER_AUTO_DISCOVERY", "ENABLE_HEALTH_CHECKS",
	}
	
	for _, env := range envVars {
		originalEnv[env] = os.Getenv(env)
		os.Unsetenv(env)
	}
	
	// Restore environment after test
	defer func() {
		for env, value := range originalEnv {
			if value != "" {
				os.Setenv(env, value)
			} else {
				os.Unsetenv(env)
			}
		}
	}()

	t.Run("DefaultValues", func(t *testing.T) {
		cfg := config.Load()
		
		// Test default values
		if cfg.Port != "8080" {
			t.Errorf("Expected default port '8080', got '%s'", cfg.Port)
		}
		if cfg.SQLiteDBFile != "./data/ai_gateway.db" {
			t.Errorf("Expected default SQLite path './data/ai_gateway.db', got '%s'", cfg.SQLiteDBFile)
		}
		if cfg.RedisAddr != "localhost:6379" {
			t.Errorf("Expected default Redis addr 'localhost:6379', got '%s'", cfg.RedisAddr)
		}
		if cfg.LogDir != "./logs" {
			t.Errorf("Expected default log dir './logs', got '%s'", cfg.LogDir)
		}
		if cfg.LogLevel != "info" {
			t.Errorf("Expected default log level 'info', got '%s'", cfg.LogLevel)
		}
		if cfg.MaxSessions != 100 {
			t.Errorf("Expected default max sessions 100, got %d", cfg.MaxSessions)
		}
		if cfg.SessionTimeout != 3600*time.Second {
			t.Errorf("Expected default session timeout 3600s, got %v", cfg.SessionTimeout)
		}
		if cfg.WebSocketTimeout != 7200*time.Second {
			t.Errorf("Expected default websocket timeout 7200s, got %v", cfg.WebSocketTimeout)
		}
		if cfg.ClaudeCLIPath != "claude" {
			t.Errorf("Expected default Claude CLI path 'claude', got '%s'", cfg.ClaudeCLIPath)
		}
		if cfg.GeminiCLIPath != "gemini" {
			t.Errorf("Expected default Gemini CLI path 'gemini', got '%s'", cfg.GeminiCLIPath)
		}
		if !cfg.EnableProviderAutoDiscovery {
			t.Error("Expected default provider auto discovery to be true")
		}
		if !cfg.EnableHealthChecks {
			t.Error("Expected default health checks to be true")
		}
	})

	t.Run("CustomValues", func(t *testing.T) {
		// Set custom environment variables
		os.Setenv("PORT", "9090")
		os.Setenv("SQLITE_DB_FILE", "./custom/db.sqlite")
		os.Setenv("REDIS_ADDR", "redis:6379")
		os.Setenv("LOG_DIR", "./custom/logs")
		os.Setenv("LOG_LEVEL", "debug")
		os.Setenv("MAX_SESSIONS", "50")
		os.Setenv("SESSION_TIMEOUT", "1800")
		os.Setenv("WEBSOCKET_TIMEOUT", "3600")
		os.Setenv("CLAUDE_CLI_PATH", "/usr/local/bin/claude")
		os.Setenv("GEMINI_CLI_PATH", "/usr/local/bin/gemini")
		os.Setenv("ENABLE_PROVIDER_AUTO_DISCOVERY", "false")
		os.Setenv("ENABLE_HEALTH_CHECKS", "false")
		
		cfg := config.Load()
		
		// Test custom values
		if cfg.Port != "9090" {
			t.Errorf("Expected custom port '9090', got '%s'", cfg.Port)
		}
		if cfg.SQLiteDBFile != "./custom/db.sqlite" {
			t.Errorf("Expected custom SQLite path './custom/db.sqlite', got '%s'", cfg.SQLiteDBFile)
		}
		if cfg.RedisAddr != "redis:6379" {
			t.Errorf("Expected custom Redis addr 'redis:6379', got '%s'", cfg.RedisAddr)
		}
		if cfg.LogDir != "./custom/logs" {
			t.Errorf("Expected custom log dir './custom/logs', got '%s'", cfg.LogDir)
		}
		if cfg.LogLevel != "debug" {
			t.Errorf("Expected custom log level 'debug', got '%s'", cfg.LogLevel)
		}
		if cfg.MaxSessions != 50 {
			t.Errorf("Expected custom max sessions 50, got %d", cfg.MaxSessions)
		}
		if cfg.SessionTimeout != 1800*time.Second {
			t.Errorf("Expected custom session timeout 1800s, got %v", cfg.SessionTimeout)
		}
		if cfg.WebSocketTimeout != 3600*time.Second {
			t.Errorf("Expected custom websocket timeout 3600s, got %v", cfg.WebSocketTimeout)
		}
		if cfg.ClaudeCLIPath != "/usr/local/bin/claude" {
			t.Errorf("Expected custom Claude CLI path '/usr/local/bin/claude', got '%s'", cfg.ClaudeCLIPath)
		}
		if cfg.GeminiCLIPath != "/usr/local/bin/gemini" {
			t.Errorf("Expected custom Gemini CLI path '/usr/local/bin/gemini', got '%s'", cfg.GeminiCLIPath)
		}
		if cfg.EnableProviderAutoDiscovery {
			t.Error("Expected custom provider auto discovery to be false")
		}
		if cfg.EnableHealthChecks {
			t.Error("Expected custom health checks to be false")
		}
	})

	t.Run("InvalidIntegerValues", func(t *testing.T) {
		// Test invalid integer values fall back to defaults
		os.Setenv("MAX_SESSIONS", "invalid")
		os.Setenv("SESSION_TIMEOUT", "not_a_number")
		os.Setenv("WEBSOCKET_TIMEOUT", "also_invalid")
		
		cfg := config.Load()
		
		if cfg.MaxSessions != 100 {
			t.Errorf("Expected default max sessions 100 for invalid value, got %d", cfg.MaxSessions)
		}
		if cfg.SessionTimeout != 3600*time.Second {
			t.Errorf("Expected default session timeout 3600s for invalid value, got %v", cfg.SessionTimeout)
		}
		if cfg.WebSocketTimeout != 7200*time.Second {
			t.Errorf("Expected default websocket timeout 7200s for invalid value, got %v", cfg.WebSocketTimeout)
		}
	})

	t.Run("InvalidBooleanValues", func(t *testing.T) {
		// Test invalid boolean values fall back to defaults
		os.Setenv("ENABLE_PROVIDER_AUTO_DISCOVERY", "maybe")
		os.Setenv("ENABLE_HEALTH_CHECKS", "sometimes")
		
		cfg := config.Load()
		
		if !cfg.EnableProviderAutoDiscovery {
			t.Error("Expected default provider auto discovery true for invalid value")
		}
		if !cfg.EnableHealthChecks {
			t.Error("Expected default health checks true for invalid value")
		}
	})
}