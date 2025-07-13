package handlers

import (
	"net/http"
	"strconv"

	"ai-gateway-hub/internal/services"
	"ai-gateway-hub/internal/utils"

	"github.com/gin-gonic/gin"
)

// ChatHandler handles the chat page
func ChatHandler(chatService *services.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := GetLang(c)
		t := GetTranslator(c)
		chatIDStr := c.Param("id")
		utils.Debug("ChatHandler: accessing chat ID %s", chatIDStr)
		
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			utils.Error("ChatHandler: invalid chat ID %s: %v", chatIDStr, err)
			c.HTML(http.StatusBadRequest, "error.html", gin.H{
				"error": t("error.invalidChatId"),
				"lang":  lang,
			})
			return
		}

		// Get chat details
		chat, err := chatService.GetChat(chatID)
		if err != nil {
			utils.Error("ChatHandler: failed to get chat %d: %v", chatID, err)
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"error": t("error.chatNotFound"),
				"lang":  lang,
			})
			return
		}
		utils.Debug("ChatHandler: found chat %d: %s", chatID, chat.Title)

		// Get messages
		messages, err := chatService.GetMessages(chatID, 1000, 0)
		if err != nil {
			utils.Error("ChatHandler: failed to get messages for chat %d: %v", chatID, err)
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"error": t("error.failedToLoadMessages"),
				"lang":  lang,
			})
			return
		}
		utils.Debug("ChatHandler: found %d messages for chat %d", len(messages), chatID)

		utils.Debug("ChatHandler: rendering chat.html template")
		c.HTML(http.StatusOK, "chat.html", gin.H{
			"title":    chat.Title,
			"chat":     chat,
			"messages": messages,
			"lang":     lang,
		})
	}
}