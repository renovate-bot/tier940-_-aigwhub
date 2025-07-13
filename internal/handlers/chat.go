package handlers

import (
	"net/http"
	"strconv"

	"ai-gateway-hub/internal/services"

	"github.com/gin-gonic/gin"
)

// ChatHandler handles the chat page
func ChatHandler(chatService *services.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := GetTranslator(c)
		chatIDStr := c.Param("id")
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{
				"error": t("error.invalidChatId"),
				"t":     t,
				"lang":  GetLang(c),
			})
			return
		}

		// Get chat details
		chat, err := chatService.GetChat(chatID)
		if err != nil {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"error": t("error.chatNotFound"),
				"t":     t,
				"lang":  GetLang(c),
			})
			return
		}

		// Get messages
		messages, err := chatService.GetMessages(chatID, 1000, 0)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"error": t("error.failedToLoadMessages"),
				"t":     t,
				"lang":  GetLang(c),
			})
			return
		}

		c.HTML(http.StatusOK, "chat.html", gin.H{
			"title":    chat.Title,
			"chat":     chat,
			"messages": messages,
			"t":        t,
			"lang":     GetLang(c),
		})
	}
}