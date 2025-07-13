package handlers

import (
	"ai-gateway-hub/internal/i18n"

	"github.com/gin-gonic/gin"
)

// GetLang extracts the language from the context
func GetLang(c *gin.Context) string {
	if lang, exists := c.Get("lang"); exists {
		if langStr, ok := lang.(string); ok {
			return langStr
		}
	}
	return "en"
}

// GetTranslator returns a translation function for templates
func GetTranslator(c *gin.Context) func(string, ...interface{}) string {
	lang := GetLang(c)
	return func(key string, args ...interface{}) string {
		return i18n.T(lang, key, args...)
	}
}