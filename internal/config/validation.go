package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// ValidationResult holds the result of configuration validation
type ValidationResult struct {
	Valid   bool
	Errors  []string
	Warnings []string
}

// Validate validates the configuration and returns validation results
func (c *Config) Validate() *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	// Validate port
	if err := c.validatePort(); err != nil {
		result.addError(err.Error())
	}

	// Validate directories
	c.validateDirectories(result)

	// Validate timeouts
	c.validateTimeouts(result)

	// Validate CLI paths
	c.validateCLIPaths(result)

	// Validate feature flags
	c.validateFeatureFlags(result)

	// Set overall validity
	result.Valid = len(result.Errors) == 0

	return result
}

// validatePort validates the port configuration
func (c *Config) validatePort() error {
	if c.Port == "" {
		return fmt.Errorf("PORT is required")
	}

	port, err := strconv.Atoi(c.Port)
	if err != nil {
		return fmt.Errorf("PORT must be a valid number, got: %s", c.Port)
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("PORT must be between 1 and 65535, got: %d", port)
	}

	if port < 1024 {
		// This is a warning, not an error
		return nil
	}

	return nil
}

// validateDirectories validates directory configurations
func (c *Config) validateDirectories(result *ValidationResult) {
	directories := map[string]string{
		"STATIC_DIR":   c.StaticDir,
		"TEMPLATE_DIR": c.TemplateDir,
		"LOG_DIR":      c.LogDir,
	}

	for name, path := range directories {
		if path == "" {
			result.addError(fmt.Sprintf("%s is required", name))
			continue
		}

		// Check if directory exists or can be created
		if err := c.ensureDirectoryExists(path); err != nil {
			result.addError(fmt.Sprintf("%s (%s): %v", name, path, err))
		}
	}

	// Validate database file directory
	if c.SQLiteDBFile != "" {
		dbDir := filepath.Dir(c.SQLiteDBFile)
		if err := c.ensureDirectoryExists(dbDir); err != nil {
			result.addError(fmt.Sprintf("SQLITE_DB_FILE directory (%s): %v", dbDir, err))
		}
	}
}

// validateTimeouts validates timeout configurations
func (c *Config) validateTimeouts(result *ValidationResult) {
	if c.SessionTimeout <= 0 {
		result.addError("SESSION_TIMEOUT must be positive")
	}

	if c.WebSocketTimeout <= 0 {
		result.addError("WEBSOCKET_TIMEOUT must be positive")
	}

	if c.SessionTimeout > 24*time.Hour {
		result.addWarning("SESSION_TIMEOUT is very long (>24h), consider shorter duration for security")
	}

	if c.WebSocketTimeout > 12*time.Hour {
		result.addWarning("WEBSOCKET_TIMEOUT is very long (>12h), consider shorter duration")
	}
}

// validateCLIPaths validates AI provider CLI paths
func (c *Config) validateCLIPaths(result *ValidationResult) {
	if c.ClaudeCLIPath == "" {
		result.addWarning("CLAUDE_CLI_PATH is empty, Claude provider will be unavailable")
	} else if !c.isExecutableAvailable(c.ClaudeCLIPath) {
		result.addWarning(fmt.Sprintf("Claude CLI not found at path: %s", c.ClaudeCLIPath))
	}

	if c.GeminiCLIPath == "" {
		result.addWarning("GEMINI_CLI_PATH is empty, Gemini provider will be unavailable")
	} else if !c.isExecutableAvailable(c.GeminiCLIPath) {
		result.addWarning(fmt.Sprintf("Gemini CLI not found at path: %s", c.GeminiCLIPath))
	}
}

// validateFeatureFlags validates feature flag configurations
func (c *Config) validateFeatureFlags(result *ValidationResult) {
	if c.MaxSessions <= 0 {
		result.addError("MAX_SESSIONS must be positive")
	}

	if c.MaxSessions > 10000 {
		result.addWarning("MAX_SESSIONS is very high (>10000), may impact performance")
	}
}

// ensureDirectoryExists checks if directory exists and creates it if needed
func (c *Config) ensureDirectoryExists(path string) error {
	if path == "" {
		return fmt.Errorf("path is empty")
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist, try to create it
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("cannot create directory: %w", err)
			}
			return nil
		}
		return fmt.Errorf("cannot access directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path exists but is not a directory")
	}

	return nil
}

// isExecutableAvailable checks if an executable is available in PATH or as absolute path
func (c *Config) isExecutableAvailable(path string) bool {
	if filepath.IsAbs(path) {
		// Absolute path - check if file exists and is executable
		info, err := os.Stat(path)
		if err != nil {
			return false
		}
		return !info.IsDir() && info.Mode()&0111 != 0
	}

	// Relative path - check in PATH
	_, err := exec.LookPath(path)
	return err == nil
}

// addError adds an error to the validation result
func (r *ValidationResult) addError(message string) {
	r.Errors = append(r.Errors, message)
}

// addWarning adds a warning to the validation result
func (r *ValidationResult) addWarning(message string) {
	r.Warnings = append(r.Warnings, message)
}

// HasErrors returns true if there are validation errors
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings returns true if there are validation warnings
func (r *ValidationResult) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// Summary returns a formatted summary of validation results
func (r *ValidationResult) Summary() string {
	if r.Valid && !r.HasWarnings() {
		return "Configuration validation passed with no issues"
	}

	summary := ""
	if len(r.Errors) > 0 {
		summary += fmt.Sprintf("ERRORS (%d):\n", len(r.Errors))
		for i, err := range r.Errors {
			summary += fmt.Sprintf("  %d. %s\n", i+1, err)
		}
	}

	if len(r.Warnings) > 0 {
		if summary != "" {
			summary += "\n"
		}
		summary += fmt.Sprintf("WARNINGS (%d):\n", len(r.Warnings))
		for i, warning := range r.Warnings {
			summary += fmt.Sprintf("  %d. %s\n", i+1, warning)
		}
	}

	return summary
}