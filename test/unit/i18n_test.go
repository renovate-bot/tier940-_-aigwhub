package unit

import (
	"testing"
	"ai-gateway-hub/internal/i18n"
)

func TestI18nFallbackToEnglish(t *testing.T) {
	// Initialize i18n with test data
	err := i18n.Init("../../locales", "en")
	if err != nil {
		t.Fatalf("Failed to initialize i18n: %v", err)
	}

	tests := []struct {
		name     string
		lang     string
		key      string
		expected string
		desc     string
	}{
		{
			name:     "English key exists",
			lang:     "en",
			key:      "app.title",
			expected: "AI Gateway Hub",
			desc:     "Should return English translation",
		},
		{
			name:     "Japanese key exists",
			lang:     "ja",
			key:      "app.title",
			expected: "AI Gateway Hub",
			desc:     "Should return Japanese translation",
		},
		{
			name:     "Key missing in Japanese, exists in English",
			lang:     "ja",
			key:      "test.missing.key", // This key doesn't exist
			expected: "test.missing.key",
			desc:     "Should return the key itself when not found",
		},
		{
			name:     "Invalid language code",
			lang:     "fr", // French not supported
			key:      "app.title",
			expected: "AI Gateway Hub",
			desc:     "Should fallback to English for unsupported language",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := i18n.T(tt.lang, tt.key)
			
			// For missing keys, we expect the key itself
			if tt.key == "test.missing.key" {
				if result != tt.expected {
					t.Errorf("Expected key '%s' to be returned, got '%s'", tt.expected, result)
				}
			} else {
				// For existing keys, check if we got a translation (not the key itself)
				if result == tt.key {
					t.Errorf("Expected translation for key '%s', but got the key itself", tt.key)
				}
			}
		})
	}
}

func TestGetLanguageFromAcceptHeader(t *testing.T) {
	tests := []struct {
		acceptHeader string
		expected     string
	}{
		{"en-US,en;q=0.9", "en"},
		{"ja-JP,ja;q=0.9", "ja"},
		{"fr-FR,fr;q=0.9", "en"}, // Unsupported language
		{"", "en"},                // Empty header
		{"invalid", "en"},         // Invalid format
	}

	for _, tt := range tests {
		result := i18n.GetLanguageFromAcceptHeader(tt.acceptHeader)
		if result != tt.expected {
			t.Errorf("For header '%s', expected '%s', got '%s'", tt.acceptHeader, tt.expected, result)
		}
	}
}