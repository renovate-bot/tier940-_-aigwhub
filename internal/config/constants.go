package config

// Application constants
const (
	// Default values
	DefaultLanguage = "en"
	DefaultTheme    = "light"
)

// Supported values
var (
	// SupportedLanguages defines the languages supported by the application
	SupportedLanguages = []string{"en", "ja"}
	
	// SupportedThemes defines the themes supported by the application
	SupportedThemes = []string{"light", "dark", "auto"}
)

// IsValidLanguage checks if the given language is supported
func IsValidLanguage(lang string) bool {
	for _, supported := range SupportedLanguages {
		if lang == supported {
			return true
		}
	}
	return false
}

// IsValidTheme checks if the given theme is supported
func IsValidTheme(theme string) bool {
	for _, supported := range SupportedThemes {
		if theme == supported {
			return true
		}
	}
	return false
}