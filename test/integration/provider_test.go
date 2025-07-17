package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ai-gateway-hub/internal/config"
	"ai-gateway-hub/internal/providers"
	"ai-gateway-hub/internal/services"
	"ai-gateway-hub/internal/utils"
)

func TestClaudeProvider(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "claude_provider_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup path manager
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)
	utils.InitPathManager()

	// Create test log directory
	logDir := "./test_logs"
	
	t.Run("NewClaudeProvider", func(t *testing.T) {
		provider := providers.NewClaudeProvider("claude", logDir, false, "")
		
		if provider == nil {
			t.Fatal("NewClaudeProvider returned nil")
		}
		
		if provider.GetID() != "claude" {
			t.Errorf("Expected ID 'claude', got '%s'", provider.GetID())
		}
		
		if provider.GetName() != "Claude Code" {
			t.Errorf("Expected name 'Claude Code', got '%s'", provider.GetName())
		}
		
		if provider.GetDescription() == "" {
			t.Error("Description should not be empty")
		}
	})

	t.Run("IsAvailable", func(t *testing.T) {
		provider := providers.NewClaudeProvider("claude", logDir, false, "")
		
		// Note: This test will check if claude CLI is available
		// In a real environment, this should return true if claude CLI is installed
		available := provider.IsAvailable()
		t.Logf("Claude CLI available: %v", available)
		
		// We don't assert true/false here because it depends on the environment
		// but we test that the method doesn't panic
	})

	t.Run("IsAvailable_InvalidCommand", func(t *testing.T) {
		provider := providers.NewClaudeProvider("non_existent_command", logDir, false, "")
		
		available := provider.IsAvailable()
		if available {
			t.Error("Expected false for non-existent command")
		}
	})

	t.Run("SendPrompt_CreatesLogFile", func(t *testing.T) {
		provider := providers.NewClaudeProvider("echo", logDir, false, "") // Use echo instead of claude for testing
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		// This will likely fail with claude CLI, but should create log file
		response, err := provider.SendPrompt(ctx, "Hello test", 123)
		
		// Check if log file was created
		expectedLogPath := filepath.Join(tempDir, logDir, "claude", "chat_123.log")
		if _, statErr := os.Stat(expectedLogPath); statErr != nil {
			t.Errorf("Log file was not created at %s: %v", expectedLogPath, statErr)
		} else {
			// Check log file content
			content, readErr := os.ReadFile(expectedLogPath)
			if readErr != nil {
				t.Errorf("Failed to read log file: %v", readErr)
			} else {
				logContent := string(content)
				if logContent == "" {
					t.Error("Log file is empty")
				}
				t.Logf("Log content: %s", logContent)
			}
		}
		
		if response != nil {
			response.Close()
		}
		
		// We don't assert success/failure of the command itself
		// because it depends on claude CLI being available and authenticated
		t.Logf("SendPrompt result - error: %v", err)
	})

	t.Run("SendPrompt_ContextTimeout", func(t *testing.T) {
		provider := providers.NewClaudeProvider("sleep", logDir, false, "") // Use sleep command for timeout test
		
		// Very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		
		response, err := provider.SendPrompt(ctx, "5", 124) // sleep 5 seconds
		
		if response != nil {
			response.Close()
		}
		
		// Should timeout or create log file
		expectedLogPath := filepath.Join(tempDir, logDir, "claude", "chat_124.log")
		if _, statErr := os.Stat(expectedLogPath); statErr != nil {
			t.Logf("Log file creation result: %v", statErr)
		}
		
		t.Logf("Timeout test result - error: %v", err)
	})

	t.Run("StreamResponse_CreatesLogFile", func(t *testing.T) {
		provider := providers.NewClaudeProvider("echo", logDir, false, "")
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		// Create a simple writer
		var output []byte
		writer := &testWriter{data: &output}
		
		err := provider.StreamResponse(ctx, "Hello stream test", 125, writer)
		
		// Check if log file was created
		expectedLogPath := filepath.Join(tempDir, logDir, "claude", "chat_125.log")
		if _, statErr := os.Stat(expectedLogPath); statErr != nil {
			t.Errorf("Log file was not created at %s: %v", expectedLogPath, statErr)
		}
		
		t.Logf("StreamResponse result - error: %v", err)
		t.Logf("Writer received: %s", string(output))
	})
}

// testWriter implements io.Writer for testing
type testWriter struct {
	data *[]byte
}

func (w *testWriter) Write(p []byte) (int, error) {
	*w.data = append(*w.data, p...)
	return len(p), nil
}

func TestProviderRegistry(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "provider_registry_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)
	utils.InitPathManager()

	t.Run("RegisterAndGet", func(t *testing.T) {
		registry := services.NewProviderRegistry()
		provider := providers.NewClaudeProvider("test-claude", "./logs", false, "")
		
		err := registry.Register(provider)
		if err != nil {
			t.Fatalf("Failed to register provider: %v", err)
		}
		
		retrieved, err := registry.Get("claude")
		if err != nil {
			t.Fatalf("Failed to get provider: %v", err)
		}
		
		if retrieved.GetID() != "claude" {
			t.Errorf("Expected ID 'claude', got '%s'", retrieved.GetID())
		}
	})

	t.Run("RegisterDuplicate", func(t *testing.T) {
		registry := services.NewProviderRegistry()
		provider1 := providers.NewClaudeProvider("duplicate", "./logs", false, "")
		provider2 := providers.NewClaudeProvider("duplicate", "./logs", false, "")
		
		err := registry.Register(provider1)
		if err != nil {
			t.Fatalf("Failed to register first provider: %v", err)
		}
		
		err = registry.Register(provider2)
		if err == nil {
			t.Error("Expected error when registering duplicate provider")
		}
	})

	t.Run("GetNonExistent", func(t *testing.T) {
		registry := services.NewProviderRegistry()
		
		_, err := registry.Get("non-existent")
		if err == nil {
			t.Error("Expected error when getting non-existent provider")
		}
	})

	t.Run("List", func(t *testing.T) {
		registry := services.NewProviderRegistry()
		provider := providers.NewClaudeProvider("claude", "./logs", false, "")
		
		err := registry.Register(provider)
		if err != nil {
			t.Fatalf("Failed to register provider: %v", err)
		}
		
		providers := registry.List()
		if len(providers) != 1 {
			t.Errorf("Expected 1 provider, got %d", len(providers))
		}
		
		// Check that the provider is listed
		if len(providers) > 0 && providers[0].ID != "claude" {
			t.Errorf("Expected provider ID 'claude', got '%s'", providers[0].ID)
		}
	})

	t.Run("RegisterDefaultProviders", func(t *testing.T) {
		registry := services.NewProviderRegistry()
		
		cfg := &config.Config{
			LogDir:                "./test_logs",
			ClaudeCLIPath:         "claude",
			ClaudeSkipPermissions: false,
			ClaudeExtraArgs:       "",
		}
		err := registry.RegisterDefaultProviders(cfg)
		if err != nil {
			t.Fatalf("Failed to register default providers: %v", err)
		}
		
		// Should have registered Claude provider
		_, err = registry.Get("claude")
		if err != nil {
			t.Errorf("Default Claude provider not registered: %v", err)
		}
		
		providers := registry.List()
		if len(providers) == 0 {
			t.Error("No default providers were registered")
		}
	})
}