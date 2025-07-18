package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Environment represents the application environment
type Environment string

const (
	Development Environment = "development"
	Testing     Environment = "testing"
	Staging     Environment = "staging"
	Production  Environment = "production"
)

// EnvironmentConfig holds environment-specific configuration overrides
type EnvironmentConfig struct {
	Environment Environment
	Overrides   map[string]interface{}
}

// GetCurrentEnvironment determines the current environment
func GetCurrentEnvironment() Environment {
	env := strings.ToLower(viper.GetString("ENVIRONMENT"))
	if env == "" {
		env = strings.ToLower(viper.GetString("GIN_MODE"))
	}
	if env == "" {
		env = strings.ToLower(viper.GetString("NODE_ENV"))
	}

	switch env {
	case "production", "prod", "release":
		return Production
	case "staging", "stage":
		return Staging
	case "testing", "test":
		return Testing
	case "development", "dev", "debug":
		return Development
	default:
		return Development // Default to development
	}
}

// LoadWithEnvironment loads configuration with environment-specific overrides
func LoadWithEnvironment() *Config {
	config := Load()
	env := GetCurrentEnvironment()
	
	// Apply environment-specific configurations
	switch env {
	case Development:
		applyDevelopmentConfig(config)
	case Testing:
		applyTestingConfig(config)
	case Staging:
		applyStagingConfig(config)
	case Production:
		applyProductionConfig(config)
	}

	return config
}

// applyDevelopmentConfig applies development-specific settings
func applyDevelopmentConfig(config *Config) {
	// Only enable debug logging if no specific log level was set
	// This preserves explicitly configured log levels
	if config.LogLevel == "" {
		config.LogLevel = "debug"
	}

	// Shorter timeouts for faster development iteration
	if config.SessionTimeout == 3600*time.Second {
		config.SessionTimeout = 1800 * time.Second // 30 minutes
	}

	// Enable all feature flags in development
	config.EnableProviderAutoDiscovery = true
	config.EnableHealthChecks = true
}

// applyTestingConfig applies testing-specific settings
func applyTestingConfig(config *Config) {
	// Use in-memory or test-specific database
	if strings.Contains(config.SQLiteDBFile, "ai_gateway.db") {
		config.SQLiteDBFile = ":memory:" // In-memory database for tests
	}

	// Use different Redis database for tests
	if config.RedisAddr == "localhost:6379" {
		config.RedisAddr = "localhost:6379/1" // Use database 1 for tests
	}

	// Short timeouts for faster tests
	config.SessionTimeout = 60 * time.Second
	config.WebSocketTimeout = 120 * time.Second

	// Disable some features that might interfere with tests
	config.EnableProviderAutoDiscovery = false
	config.EnableHealthChecks = false

	// Reduce session limits for tests
	config.MaxSessions = 10
}

// applyStagingConfig applies staging-specific settings
func applyStagingConfig(config *Config) {
	// Only enable debug logging if no specific log level was set in staging
	// This preserves explicitly configured log levels
	if config.LogLevel == "" {
		config.LogLevel = "debug"
	}

	// Slightly more relaxed timeouts than production
	if config.SessionTimeout == 3600*time.Second {
		config.SessionTimeout = 7200 * time.Second // 2 hours
	}

	// Enable all features for staging testing
	config.EnableProviderAutoDiscovery = true
	config.EnableHealthChecks = true
}

// applyProductionConfig applies production-specific settings
func applyProductionConfig(config *Config) {
	// Ensure production-safe logging
	if config.LogLevel == "debug" {
		config.LogLevel = "info"
	}

	// Production timeouts should be reasonable
	if config.SessionTimeout > 24*time.Hour {
		config.SessionTimeout = 24 * time.Hour // Max 24 hours
	}

	if config.WebSocketTimeout > 12*time.Hour {
		config.WebSocketTimeout = 12 * time.Hour // Max 12 hours
	}

	// Conservative session limits for production
	if config.MaxSessions > 1000 {
		config.MaxSessions = 1000
	}

	// All features should be explicitly configured in production
	// Don't override feature flags set by environment variables
}

// GetEnvironmentInfo returns information about the current environment
func GetEnvironmentInfo() map[string]interface{} {
	env := GetCurrentEnvironment()
	
	return map[string]interface{}{
		"environment":    string(env),
		"is_production":  env == Production,
		"is_development": env == Development,
		"is_testing":     env == Testing,
		"is_staging":     env == Staging,
		"gin_mode":       viper.GetString("GIN_MODE"),
		"node_env":       viper.GetString("NODE_ENV"),
	}
}

// ValidateEnvironment validates environment-specific configuration
func ValidateEnvironment(config *Config) *ValidationResult {
	result := config.Validate()
	env := GetCurrentEnvironment()

	switch env {
	case Production:
		validateProductionEnvironment(config, result)
	case Staging:
		validateStagingEnvironment(config, result)
	case Testing:
		validateTestingEnvironment(config, result)
	case Development:
		validateDevelopmentEnvironment(config, result)
	}

	return result
}

// validateProductionEnvironment adds production-specific validations
func validateProductionEnvironment(config *Config, result *ValidationResult) {
	// Production should have secure settings
	if config.LogLevel == "debug" {
		result.addWarning("Debug logging enabled in production - consider using 'info' or 'warn'")
	}

	if config.SessionTimeout > 24*time.Hour {
		result.addWarning("Very long session timeout in production - consider shorter duration for security")
	}

	// Check for development-like settings
	if config.Port == "3000" || config.Port == "8080" {
		result.addWarning("Using common development port in production")
	}

	if strings.Contains(config.SQLiteDBFile, "test") || strings.Contains(config.SQLiteDBFile, "dev") {
		result.addError("Database file path suggests non-production database")
	}
}

// validateStagingEnvironment adds staging-specific validations
func validateStagingEnvironment(config *Config, result *ValidationResult) {
	// Staging should be similar to production
	if strings.Contains(config.SQLiteDBFile, "prod") {
		result.addError("Staging environment appears to use production database")
	}
}

// validateTestingEnvironment adds testing-specific validations
func validateTestingEnvironment(config *Config, result *ValidationResult) {
	// Tests should use isolated resources
	if !strings.Contains(config.SQLiteDBFile, "test") && config.SQLiteDBFile != ":memory:" {
		result.addWarning("Testing environment should use test database or in-memory database")
	}

	if config.MaxSessions > 50 {
		result.addWarning("High session limit in testing environment - consider reducing for faster tests")
	}
}

// validateDevelopmentEnvironment adds development-specific validations
func validateDevelopmentEnvironment(config *Config, result *ValidationResult) {
	// Development-specific checks
	if strings.Contains(config.SQLiteDBFile, "prod") {
		result.addError("Development environment appears to use production database")
	}

	if config.EnableHealthChecks == false {
		result.addWarning("Health checks disabled in development - consider enabling for testing")
	}
}

// ConfigSummary returns a summary of the current configuration
func ConfigSummary(config *Config) string {
	env := GetCurrentEnvironment()
	
	summary := fmt.Sprintf("AI Gateway Hub Configuration Summary\n")
	summary += fmt.Sprintf("Environment: %s\n", env)
	summary += fmt.Sprintf("Port: %s\n", config.Port)
	summary += fmt.Sprintf("Database: %s\n", config.SQLiteDBFile)
	summary += fmt.Sprintf("Redis: %s\n", config.RedisAddr)
	summary += fmt.Sprintf("Log Level: %s\n", config.LogLevel)
	summary += fmt.Sprintf("Max Sessions: %d\n", config.MaxSessions)
	summary += fmt.Sprintf("Session Timeout: %v\n", config.SessionTimeout)
	summary += fmt.Sprintf("WebSocket Timeout: %v\n", config.WebSocketTimeout)
	summary += fmt.Sprintf("Claude CLI: %s\n", config.ClaudeCLIPath)
	summary += fmt.Sprintf("Gemini CLI: %s\n", config.GeminiCLIPath)
	summary += fmt.Sprintf("Features: AutoDiscovery=%t, HealthChecks=%t\n", 
		config.EnableProviderAutoDiscovery, config.EnableHealthChecks)
	
	return summary
}