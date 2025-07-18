package middleware

import (
	"ai-gateway-hub/internal/config"
	"ai-gateway-hub/internal/i18n"

	"github.com/gin-gonic/gin"
)

// I18nMiddleware adds language detection and template functions
func I18nMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Priority order: query parameter > cookie > Accept-Language header
		lang := c.Query("lang")
		if lang == "" {
			// Check for language preference cookie
			if cookieLang, err := c.Cookie("lang"); err == nil && cookieLang != "" {
				lang = cookieLang
			}
		}
		if lang == "" {
			// Use default language from configuration
			lang = config.DefaultLanguage
		}
		
		// Store language in context
		c.Set("lang", lang)
		
		// Add template function for translations
		if tmplFuncs, exists := c.Get("templateFuncs"); exists {
			if funcs, ok := tmplFuncs.(gin.H); ok {
				funcs["t"] = func(key string, args ...interface{}) string {
					return i18n.T(lang, key, args...)
				}
			}
		} else {
			c.Set("templateFuncs", gin.H{
				"t": func(key string, args ...interface{}) string {
					return i18n.T(lang, key, args...)
				},
				"lang": lang,
			})
		}
		
		c.Next()
	}
}