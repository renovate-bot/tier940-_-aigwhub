package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SettingsHandler handles the settings page
func SettingsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := GetLang(c)

		c.HTML(http.StatusOK, "pages/settings.html", gin.H{
			"lang": lang,
		})
	}
}