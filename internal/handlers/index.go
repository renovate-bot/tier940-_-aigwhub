package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// IndexHandler handles the home page
func IndexHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := GetLang(c)
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "AI Gateway Hub", // Will be translated in template using T function
			"lang":  lang,
		})
	}
}