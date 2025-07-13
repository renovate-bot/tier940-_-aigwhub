package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// IndexHandler handles the home page
func IndexHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := GetTranslator(c)
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": t("app.title"),
			"t":     t,
			"lang":  GetLang(c),
		})
	}
}