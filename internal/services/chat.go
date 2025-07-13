package services

import (
	"database/sql"
	"fmt"
	"time"

	"ai-gateway-hub/internal/models"
)

// ChatService handles chat-related operations
type ChatService struct {
	db *sql.DB
}

func NewChatService(db *sql.DB) *ChatService {
	return &ChatService{db: db}
}

// CreateChat creates a new chat
func (s *ChatService) CreateChat(title, provider string) (*models.Chat, error) {
	query := `
		INSERT INTO chats (title, provider, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		RETURNING id, title, provider, created_at, updated_at
	`
	
	now := time.Now()
	var chat models.Chat
	
	err := s.db.QueryRow(query, title, provider, now, now).Scan(
		&chat.ID,
		&chat.Title,
		&chat.Provider,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create chat: %w", err)
	}
	
	return &chat, nil
}

// GetChat retrieves a chat by ID
func (s *ChatService) GetChat(id int64) (*models.Chat, error) {
	query := `
		SELECT id, title, provider, created_at, updated_at
		FROM chats
		WHERE id = ?
	`
	
	var chat models.Chat
	err := s.db.QueryRow(query, id).Scan(
		&chat.ID,
		&chat.Title,
		&chat.Provider,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("chat not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}
	
	return &chat, nil
}

// GetChats retrieves all chats
func (s *ChatService) GetChats(limit, offset int) ([]*models.Chat, error) {
	query := `
		SELECT id, title, provider, created_at, updated_at
		FROM chats
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get chats: %w", err)
	}
	defer rows.Close()
	
	var chats []*models.Chat
	for rows.Next() {
		var chat models.Chat
		err := rows.Scan(
			&chat.ID,
			&chat.Title,
			&chat.Provider,
			&chat.CreatedAt,
			&chat.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat: %w", err)
		}
		chats = append(chats, &chat)
	}
	
	return chats, nil
}

// UpdateChat updates a chat's details
func (s *ChatService) UpdateChat(id int64, title string) error {
	query := `
		UPDATE chats
		SET title = ?, updated_at = ?
		WHERE id = ?
	`
	
	_, err := s.db.Exec(query, title, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update chat: %w", err)
	}
	
	return nil
}

// DeleteChat deletes a chat and its messages
func (s *ChatService) DeleteChat(id int64) error {
	query := `DELETE FROM chats WHERE id = ?`
	
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}
	
	return nil
}

// AddMessage adds a message to a chat
func (s *ChatService) AddMessage(chatID int64, role, content string) (*models.Message, error) {
	// Update chat's updated_at timestamp
	updateQuery := `UPDATE chats SET updated_at = ? WHERE id = ?`
	if _, err := s.db.Exec(updateQuery, time.Now(), chatID); err != nil {
		return nil, fmt.Errorf("failed to update chat timestamp: %w", err)
	}
	
	// Insert message
	query := `
		INSERT INTO messages (chat_id, role, content, created_at)
		VALUES (?, ?, ?, ?)
		RETURNING id, chat_id, role, content, created_at
	`
	
	var msg models.Message
	err := s.db.QueryRow(query, chatID, role, content, time.Now()).Scan(
		&msg.ID,
		&msg.ChatID,
		&msg.Role,
		&msg.Content,
		&msg.CreatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to add message: %w", err)
	}
	
	return &msg, nil
}

// GetMessages retrieves messages for a chat
func (s *ChatService) GetMessages(chatID int64, limit, offset int) ([]*models.Message, error) {
	query := `
		SELECT id, chat_id, role, content, created_at
		FROM messages
		WHERE chat_id = ?
		ORDER BY created_at ASC
		LIMIT ? OFFSET ?
	`
	
	rows, err := s.db.Query(query, chatID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()
	
	var messages []*models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(
			&msg.ID,
			&msg.ChatID,
			&msg.Role,
			&msg.Content,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, &msg)
	}
	
	return messages, nil
}