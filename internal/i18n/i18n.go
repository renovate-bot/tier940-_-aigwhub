package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
)

// Localizer handles internationalization
type Localizer struct {
	translations map[string]map[string]string
	defaultLang  string
	mu           sync.RWMutex
}

var (
	instance *Localizer
	once     sync.Once
)

// Init initializes the i18n system
func Init(localesDir string, defaultLang string) error {
	var initErr error
	once.Do(func() {
		instance = &Localizer{
			translations: make(map[string]map[string]string),
			defaultLang:  defaultLang,
		}
		initErr = instance.loadTranslations(localesDir)
	})
	return initErr
}

// InitWithFS initializes the i18n system with embedded file system
func InitWithFS(localeFS embed.FS, defaultLang string) error {
	var initErr error
	once.Do(func() {
		instance = &Localizer{
			translations: make(map[string]map[string]string),
			defaultLang:  defaultLang,
		}
		initErr = instance.loadTranslationsFS(localeFS)
	})
	return initErr
}

// Get returns the singleton localizer instance
func Get() *Localizer {
	if instance == nil {
		panic("i18n not initialized. Call Init() first")
	}
	return instance
}

// T translates a key to the specified language
func T(lang, key string, args ...interface{}) string {
	return Get().Translate(lang, key, args...)
}

// loadTranslations loads all translation files
func (l *Localizer) loadTranslations(localesDir string) error {
	languages := []string{"en", "ja"}
	
	for _, lang := range languages {
		filePath := filepath.Join(localesDir, lang, "messages.json")
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read translation file %s: %w", filePath, err)
		}
		
		var translations map[string]string
		if err := json.Unmarshal(data, &translations); err != nil {
			return fmt.Errorf("failed to parse translation file %s: %w", filePath, err)
		}
		
		l.mu.Lock()
		l.translations[lang] = translations
		l.mu.Unlock()
	}
	
	return nil
}

// loadTranslationsFS loads all translation files from embedded file system
func (l *Localizer) loadTranslationsFS(localeFS embed.FS) error {
	languages := []string{"en", "ja"}
	
	for _, lang := range languages {
		filePath := filepath.Join("locales", lang, "messages.json")
		data, err := fs.ReadFile(localeFS, filePath)
		if err != nil {
			return fmt.Errorf("failed to read translation file %s: %w", filePath, err)
		}
		
		var translations map[string]string
		if err := json.Unmarshal(data, &translations); err != nil {
			return fmt.Errorf("failed to parse translation file %s: %w", filePath, err)
		}
		
		l.mu.Lock()
		l.translations[lang] = translations
		l.mu.Unlock()
	}
	
	return nil
}

// Translate returns the translated string for the given key
func (l *Localizer) Translate(lang, key string, args ...interface{}) string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	// Use default language if specified language not found
	langTranslations, ok := l.translations[lang]
	if !ok {
		langTranslations = l.translations[l.defaultLang]
	}
	
	// Get translation
	translation, ok := langTranslations[key]
	if !ok {
		// Try default language if key not found
		if lang != l.defaultLang {
			if defaultTranslations, ok := l.translations[l.defaultLang]; ok {
				if defaultTranslation, ok := defaultTranslations[key]; ok {
					translation = defaultTranslation
				} else {
					return key // Return key if not found
				}
			}
		} else {
			return key // Return key if not found
		}
	}
	
	// Format with arguments if provided
	if len(args) > 0 {
		return fmt.Sprintf(translation, args...)
	}
	
	return translation
}

// GetLanguageFromAcceptHeader parses Accept-Language header
func GetLanguageFromAcceptHeader(acceptLang string) string {
	if acceptLang == "" {
		return "en"
	}
	
	// Simple parsing - take the first language
	parts := strings.Split(acceptLang, ",")
	if len(parts) > 0 {
		lang := strings.TrimSpace(parts[0])
		// Extract language code (e.g., "en-US" -> "en")
		if idx := strings.Index(lang, "-"); idx > 0 {
			lang = lang[:idx]
		}
		if idx := strings.Index(lang, ";"); idx > 0 {
			lang = lang[:idx]
		}
		
		// Check if we support this language
		supportedLangs := []string{"en", "ja"}
		for _, supported := range supportedLangs {
			if lang == supported {
				return lang
			}
		}
	}
	
	return "en" // Default to English
}

// Middleware returns a function to extract language from context
func Middleware() func(string) string {
	return func(acceptLang string) string {
		return GetLanguageFromAcceptHeader(acceptLang)
	}
}