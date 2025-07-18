package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"ai-gateway-hub/internal/models"
	"ai-gateway-hub/internal/services"
	"ai-gateway-hub/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	// WebSocket message size limit (512KB)
	MaxWebSocketMessageSize = 512 * 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return checkWebSocketOrigin(r)
	},
}

// checkWebSocketOrigin validates the origin of WebSocket connections
func checkWebSocketOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		utils.Warn("WebSocket connection attempted without Origin header")
		return false
	}

	// Get allowed origins from environment variable
	allowedOrigins := os.Getenv("ALLOWED_WEBSOCKET_ORIGINS")
	
	// Default to development settings if not configured
	if allowedOrigins == "" {
		// Development mode: allow localhost and 127.0.0.1
		if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") || strings.Contains(origin, "172.18.0.1") {
			return true
		}
		utils.Warn("WebSocket connection from unallowed origin in development: %s", origin)
		return false
	}

	// Production mode: check against whitelist
	origins := strings.Split(allowedOrigins, ",")
	for _, allowedOrigin := range origins {
		allowedOrigin = strings.TrimSpace(allowedOrigin)
		if origin == allowedOrigin {
			utils.Debug("WebSocket connection allowed from origin: %s", origin)
			return true
		}
	}

	utils.Warn("WebSocket connection rejected from disallowed origin: %s", origin)
	return false
}

// authenticateWebSocketRequest performs basic authentication for WebSocket connections
// This is a simple implementation - you should enhance this based on your authentication system
func authenticateWebSocketRequest(r *http.Request) bool {
	// Option 1: Check for session cookie (if you're using cookie-based sessions)
	sessionCookie, err := r.Cookie("session_id")
	if err != nil || sessionCookie.Value == "" {
		// No session cookie, check for Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// For development: allow connections without authentication
			// In production: return false
			env := os.Getenv("ENVIRONMENT")
			if env == "development" || env == "" {
				utils.Debug("WebSocket connection allowed without authentication in development mode")
				return true
			}
			utils.Warn("WebSocket connection missing authentication")
			return false
		}
		
		// TODO: Validate Authorization header (Bearer token, etc.)
		// For now, accept any Authorization header
		utils.Debug("WebSocket connection authenticated via Authorization header")
		return true
	}

	// TODO: Validate session cookie with your session store
	// For now, accept any session cookie
	utils.Debug("WebSocket connection authenticated via session cookie: %s", sessionCookie.Value[:8]+"...")
	return true
}

// Client represents a WebSocket client
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	chatID   int64
	provider string
	mu       sync.Mutex
}

// Hub maintains active WebSocket connections
type Hub struct {
	clients          map[*Client]bool
	broadcast        chan []byte
	register         chan *Client
	unregister       chan *Client
	sessionService   *services.SessionService
	chatService      *services.ChatService
	providerRegistry *services.ProviderRegistry
	mu               sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub(sessionService *services.SessionService, chatService *services.ChatService, providerRegistry *services.ProviderRegistry) *Hub {
	return &Hub{
		clients:          make(map[*Client]bool),
		broadcast:        make(chan []byte),
		register:         make(chan *Client),
		unregister:       make(chan *Client),
		sessionService:   sessionService,
		chatService:      chatService,
		providerRegistry: providerRegistry,
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			utils.Debug("WebSocket client registered: %p", client)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.mu.Unlock()
				utils.Debug("WebSocket client unregistered: %p", client)
			} else {
				h.mu.Unlock()
			}

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// WebSocketHandler handles WebSocket connections
func WebSocketHandler(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Basic authentication check - you can enhance this based on your auth system
		if !authenticateWebSocketRequest(c.Request) {
			utils.Warn("WebSocket authentication failed for %s", c.ClientIP())
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			utils.Error("WebSocket upgrade failed: %v", err)
			return
		}

		// Set message size limit for security
		conn.SetReadLimit(MaxWebSocketMessageSize) // 512KB max message size

		client := &Client{
			hub:  hub,
			conn: conn,
			send: make(chan []byte, 256),
		}

		client.hub.register <- client
		utils.Debug("WebSocket client authenticated and registered: %s", c.ClientIP())

		// Start goroutines for reading and writing
		go client.writePump()
		go client.readPump()
	}
}

// readPump handles incoming messages from the WebSocket
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				utils.Error("WebSocket error: %v", err)
			}
			break
		}

		// Parse message
		var msg models.WebSocketMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			utils.Error("Failed to parse WebSocket message: %v", err)
			continue
		}

		// Handle message based on type
		switch msg.Type {
		case "ai_prompt":
			c.handleAIPrompt(msg.Data)
		case "session_status":
			c.handleSessionStatus(msg.Data)
		default:
			utils.Warn("Unknown WebSocket message type: %s", msg.Type)
		}
	}
}

// writePump handles outgoing messages to the WebSocket
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleAIPrompt processes AI prompts
func (c *Client) handleAIPrompt(data models.WSMsgData) {
	c.mu.Lock()
	c.chatID = data.ChatID
	c.provider = data.Provider
	c.mu.Unlock()

	// Get the AI provider
	provider, err := c.hub.providerRegistry.Get(data.Provider)
	if err != nil {
		c.sendError("Provider not found: " + err.Error())
		return
	}

	// Check if provider is available
	if !provider.IsAvailable() {
		c.sendError("Provider is not available")
		return
	}

	// Save user message
	if _, err := c.hub.chatService.AddMessage(data.ChatID, "user", data.Content); err != nil {
		utils.Error("Failed to save user message: %v", err)
	}

	// Stream response
	go func() {
		// Create context for cancellation
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		
		var responseContent string
		writer := &websocketWriter{client: c, buffer: &responseContent}

		err := provider.StreamResponse(ctx, data.Content, data.ChatID, writer)
		
		// Always send completion message to indicate end of streaming
		c.sendStreamCompletion(data.ChatID)
		
		if err != nil {
			c.sendError("Failed to get response: " + err.Error())
			return
		}

		// Save assistant message
		if responseContent != "" {
			if _, err := c.hub.chatService.AddMessage(data.ChatID, "assistant", responseContent); err != nil {
				utils.Error("Failed to save assistant message: %v", err)
			}
		}
	}()
}

// handleSessionStatus handles session status updates
func (c *Client) handleSessionStatus(data models.WSMsgData) {
	// Update session with chat ID if provided
	if data.ChatID > 0 {
		c.mu.Lock()
		c.chatID = data.ChatID
		c.mu.Unlock()
	}
}

// sendError sends an error message to the client
func (c *Client) sendError(message string) {
	msg := models.WebSocketMessage{
		Type: "error",
		Data: models.WSMsgData{
			Content:   message,
			Timestamp: time.Now(),
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		utils.Error("Failed to marshal error message: %v", err)
		return
	}

	select {
	case c.send <- data:
	default:
		utils.Error("Failed to send error message to client")
	}
}

// sendStreamCompletion sends a stream completion message to the client
func (c *Client) sendStreamCompletion(chatID int64) {
	msg := models.WebSocketMessage{
		Type: "ai_response_end",
		Data: models.WSMsgData{
			ChatID:    chatID,
			Provider:  c.provider,
			Timestamp: time.Now(),
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		utils.Error("Failed to marshal stream completion message: %v", err)
		return
	}

	select {
	case c.send <- data:
		utils.Debug("Stream completion sent for chat %d", chatID)
	default:
		utils.Error("Failed to send stream completion message to client")
	}
}

// websocketWriter implements io.Writer for streaming to WebSocket
type websocketWriter struct {
	client *Client
	buffer *string
}

func (w *websocketWriter) Write(p []byte) (n int, err error) {
	content := string(p)
	*w.buffer += content

	msg := models.WebSocketMessage{
		Type: "ai_response",
		Data: models.WSMsgData{
			ChatID:    w.client.chatID,
			Provider:  w.client.provider,
			Content:   content,
			Timestamp: time.Now(),
			Stream:    true,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return 0, err
	}

	select {
	case w.client.send <- data:
		return len(p), nil
	default:
		return 0, io.ErrClosedPipe
	}
}