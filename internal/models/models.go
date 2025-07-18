package models

import (
	"database/sql/driver"
	"time"
)

// Chat represents a conversation session
type Chat struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Message represents a single message in a chat
type Message struct {
	ID        int64     `json:"id"`
	ChatID    int64     `json:"chat_id"`
	Role      string    `json:"role"` // user, assistant, system
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Session represents a WebSocket session
type Session struct {
	ID        string     `json:"id"`
	ChatID    *int64     `json:"chat_id,omitempty"`
	Data      string     `json:"data,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// WebSocketMessage represents messages sent over WebSocket
type WebSocketMessage struct {
	Type      string    `json:"type"` // ai_prompt, ai_response, session_status, error
	Data      WSMsgData `json:"data"`
}

// WSMsgData contains the actual message data
type WSMsgData struct {
	ChatID    int64     `json:"chat_id,omitempty"`
	Provider  string    `json:"provider,omitempty"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Stream    bool      `json:"stream,omitempty"`
}

// Provider represents an AI provider
type Provider struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Available   bool   `json:"available"`
	Status      string `json:"status,omitempty"`  // "ready", "not_installed", "not_configured", "error"
	Version     string `json:"version,omitempty"`
	Details     string `json:"details,omitempty"`
}

// NullTime implements sql.Scanner and driver.Valuer for nullable time fields
type NullTime struct {
	Time  time.Time
	Valid bool
}

func (nt *NullTime) Scan(value interface{}) error {
	if value == nil {
		nt.Time, nt.Valid = time.Time{}, false
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		nt.Time, nt.Valid = v, true
		return nil
	case string:
		t, err := time.Parse("2006-01-02 15:04:05", v)
		if err != nil {
			return err
		}
		nt.Time, nt.Valid = t, true
		return nil
	default:
		nt.Valid = false
		return nil
	}
}

func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}