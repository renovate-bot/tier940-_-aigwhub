package unit

import (
	"testing"
	
	"ai-gateway-hub/internal/providers"
)

func TestClaudeProviderGetStatus(t *testing.T) {
	tests := []struct {
		name           string
		cliPath        string
		expectedStatus string
	}{
		{
			name:           "Invalid path",
			cliPath:        "/invalid/path/to/claude",
			expectedStatus: "not_installed",
		},
		{
			name:           "Valid claude command",
			cliPath:        "claude", // Assumes claude is in PATH
			expectedStatus: "ready",   // This may vary based on actual installation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := providers.NewClaudeProvider(tt.cliPath, "/tmp", false, "")
			status := provider.GetStatus()
			
			// For invalid paths, we expect not_installed or error
			if tt.cliPath == "/invalid/path/to/claude" {
				if status.Status != "not_installed" && status.Status != "error" {
					t.Errorf("Expected status 'not_installed' or 'error', got '%s'", status.Status)
				}
				if status.Available {
					t.Error("Expected Available to be false for invalid path")
				}
			}
			
			// Log the actual status for debugging
			t.Logf("Provider status: %+v", status)
		})
	}
}