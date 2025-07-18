package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ai-gateway-hub/internal/database"
)

func setupTestChatService(t *testing.T) (*ChatService, func()) {
	db, err := database.InitTestDB()
	require.NoError(t, err)

	service := NewChatService(db)

	cleanup := func() {
		db.Close()
	}

	return service, cleanup
}

func TestChatService_CreateChat(t *testing.T) {
	service, cleanup := setupTestChatService(t)
	defer cleanup()

	tests := []struct {
		name     string
		title    string
		provider string
		wantErr  bool
	}{
		{
			name:     "create valid chat",
			title:    "Test Chat",
			provider: "claude",
			wantErr:  false,
		},
		{
			name:     "create chat with empty title",
			title:    "",
			provider: "claude",
			wantErr:  false,
		},
		{
			name:     "create chat with long title",
			title:    "This is a very long title that exceeds the normal length of a chat title and should be handled properly by the service",
			provider: "gemini",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chat, err := service.CreateChat(tt.title, tt.provider)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, chat)
				assert.NotZero(t, chat.ID)
				assert.Equal(t, tt.title, chat.Title)
				assert.Equal(t, tt.provider, chat.Provider)
				assert.NotZero(t, chat.CreatedAt)
				assert.NotZero(t, chat.UpdatedAt)
			}
		})
	}
}

func TestChatService_GetChat(t *testing.T) {
	service, cleanup := setupTestChatService(t)
	defer cleanup()

	// Create a test chat
	originalChat, err := service.CreateChat("Test Chat for Get", "claude")
	require.NoError(t, err)

	tests := []struct {
		name    string
		chatID  int64
		wantErr bool
	}{
		{
			name:    "get existing chat",
			chatID:  originalChat.ID,
			wantErr: false,
		},
		{
			name:    "get non-existing chat",
			chatID:  99999,
			wantErr: true,
		},
		{
			name:    "get chat with invalid ID",
			chatID:  -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chat, err := service.GetChat(tt.chatID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, originalChat.ID, chat.ID)
				assert.Equal(t, originalChat.Title, chat.Title)
				assert.Equal(t, originalChat.Provider, chat.Provider)
			}
		})
	}
}

func TestChatService_GetChats(t *testing.T) {
	service, cleanup := setupTestChatService(t)
	defer cleanup()

	// Create multiple test chats
	for i := 0; i < 5; i++ {
		_, err := service.CreateChat("Test Chat "+string(rune('A'+i)), "claude")
		require.NoError(t, err)
	}

	tests := []struct {
		name      string
		limit     int
		offset    int
		wantCount int
	}{
		{
			name:      "get all chats",
			limit:     10,
			offset:    0,
			wantCount: 5,
		},
		{
			name:      "get chats with limit",
			limit:     3,
			offset:    0,
			wantCount: 3,
		},
		{
			name:      "get chats with offset",
			limit:     10,
			offset:    3,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chats, err := service.GetChats(tt.limit, tt.offset)
			assert.NoError(t, err)
			assert.Len(t, chats, tt.wantCount)

			// Verify chats are ordered by updated_at DESC
			for i := 1; i < len(chats); i++ {
				assert.True(t, chats[i-1].UpdatedAt.After(chats[i].UpdatedAt) || chats[i-1].UpdatedAt.Equal(chats[i].UpdatedAt))
			}
		})
	}
}

func TestChatService_UpdateChat(t *testing.T) {
	service, cleanup := setupTestChatService(t)
	defer cleanup()

	// Create a test chat
	originalChat, err := service.CreateChat("Original Title", "claude")
	require.NoError(t, err)

	tests := []struct {
		name    string
		chatID  int64
		title   string
		wantErr bool
	}{
		{
			name:    "update existing chat",
			chatID:  originalChat.ID,
			title:   "Updated Title",
			wantErr: false,
		},
		{
			name:    "update non-existing chat",
			chatID:  99999,
			title:   "Non-existing",
			wantErr: false, // SQLite doesn't return error for UPDATE with no matches
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateChat(tt.chatID, tt.title)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the update for existing chat
				if tt.chatID == originalChat.ID {
					updatedChat, err := service.GetChat(tt.chatID)
					assert.NoError(t, err)
					assert.Equal(t, tt.title, updatedChat.Title)
					assert.True(t, updatedChat.UpdatedAt.After(originalChat.UpdatedAt))
				}
			}
		})
	}
}

func TestChatService_DeleteChat(t *testing.T) {
	service, cleanup := setupTestChatService(t)
	defer cleanup()

	// Create test chats
	chat1, err := service.CreateChat("Chat 1", "claude")
	require.NoError(t, err)
	chat2, err := service.CreateChat("Chat 2", "gemini")
	require.NoError(t, err)

	// Add messages to chat1
	_, err = service.AddMessage(chat1.ID, "user", "Hello")
	require.NoError(t, err)

	tests := []struct {
		name    string
		chatID  int64
		wantErr bool
		verify  func(t *testing.T)
	}{
		{
			name:    "delete existing chat with messages",
			chatID:  chat1.ID,
			wantErr: false,
			verify: func(t *testing.T) {
				// Verify chat is deleted
				_, err := service.GetChat(chat1.ID)
				assert.Error(t, err)

				// Verify messages are deleted (cascade)
				messages, err := service.GetMessages(chat1.ID, 10, 0)
				assert.NoError(t, err)
				assert.Empty(t, messages)
			},
		},
		{
			name:    "delete existing chat without messages",
			chatID:  chat2.ID,
			wantErr: false,
			verify: func(t *testing.T) {
				_, err := service.GetChat(chat2.ID)
				assert.Error(t, err)
			},
		},
		{
			name:    "delete non-existing chat",
			chatID:  99999,
			wantErr: false, // SQLite doesn't return error for DELETE with no matches
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteChat(tt.chatID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.verify != nil {
					tt.verify(t)
				}
			}
		})
	}
}

func TestChatService_AddMessage(t *testing.T) {
	service, cleanup := setupTestChatService(t)
	defer cleanup()

	// Create a test chat
	chat, err := service.CreateChat("Test Chat", "claude")
	require.NoError(t, err)

	tests := []struct {
		name    string
		chatID  int64
		role    string
		content string
		wantErr bool
	}{
		{
			name:    "add user message",
			chatID:  chat.ID,
			role:    "user",
			content: "Hello, AI!",
			wantErr: false,
		},
		{
			name:    "add assistant message",
			chatID:  chat.ID,
			role:    "assistant",
			content: "Hello! How can I help you today?",
			wantErr: false,
		},
		{
			name:    "add message with very long content",
			chatID:  chat.ID,
			role:    "user",
			content: string(make([]byte, 10000)), // 10KB message
			wantErr: false,
		},
		{
			name:    "add message to non-existing chat",
			chatID:  99999,
			role:    "user",
			content: "This should fail",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := service.AddMessage(tt.chatID, tt.role, tt.content)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, msg)
				assert.NotZero(t, msg.ID)
				assert.Equal(t, tt.chatID, msg.ChatID)
				assert.Equal(t, tt.role, msg.Role)
				assert.Equal(t, tt.content, msg.Content)
				assert.NotZero(t, msg.CreatedAt)
			}
		})
	}
}

func TestChatService_GetMessages(t *testing.T) {
	service, cleanup := setupTestChatService(t)
	defer cleanup()

	// Create test chats
	chat1, err := service.CreateChat("Chat 1", "claude")
	require.NoError(t, err)
	chat2, err := service.CreateChat("Chat 2", "gemini")
	require.NoError(t, err)

	// Add messages to chat1
	messages := []struct {
		role    string
		content string
	}{
		{"user", "First message"},
		{"assistant", "First response"},
		{"user", "Second message"},
		{"assistant", "Second response"},
	}

	for _, msg := range messages {
		_, err := service.AddMessage(chat1.ID, msg.role, msg.content)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		chatID        int64
		limit         int
		offset        int
		expectedCount int
	}{
		{
			name:          "get all messages from chat",
			chatID:        chat1.ID,
			limit:         10,
			offset:        0,
			expectedCount: 4,
		},
		{
			name:          "get messages with limit",
			chatID:        chat1.ID,
			limit:         2,
			offset:        0,
			expectedCount: 2,
		},
		{
			name:          "get messages with offset",
			chatID:        chat1.ID,
			limit:         10,
			offset:        2,
			expectedCount: 2,
		},
		{
			name:          "get messages from empty chat",
			chatID:        chat2.ID,
			limit:         10,
			offset:        0,
			expectedCount: 0,
		},
		{
			name:          "get messages from non-existing chat",
			chatID:        99999,
			limit:         10,
			offset:        0,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgs, err := service.GetMessages(tt.chatID, tt.limit, tt.offset)
			assert.NoError(t, err)
			assert.Len(t, msgs, tt.expectedCount)

			// Verify messages are ordered by created_at ASC
			for i := 1; i < len(msgs); i++ {
				assert.True(t, msgs[i].CreatedAt.After(msgs[i-1].CreatedAt) || msgs[i].CreatedAt.Equal(msgs[i-1].CreatedAt))
			}
		})
	}
}