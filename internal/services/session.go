package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ai-gateway-hub/internal/models"

	"github.com/go-redis/redis/v8"
)

// SessionService handles session management using Redis
type SessionService struct {
	redis *redis.Client
}

func NewSessionService(redisClient *redis.Client) *SessionService {
	return &SessionService{
		redis: redisClient,
	}
}

// CreateSession creates a new session
func (s *SessionService) CreateSession(sessionID string, chatID *int64, ttl time.Duration) error {
	ctx := context.Background()
	session := &models.Session{
		ID:        sessionID,
		ChatID:    chatID,
		CreatedAt: time.Now(),
	}

	if ttl > 0 {
		expiresAt := time.Now().Add(ttl)
		session.ExpiresAt = &expiresAt
	}

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	return s.redis.Set(ctx, s.key(sessionID), data, ttl).Err()
}

// GetSession retrieves a session by ID
func (s *SessionService) GetSession(sessionID string) (*models.Session, error) {
	ctx := context.Background()
	data, err := s.redis.Get(ctx, s.key(sessionID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session models.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// UpdateSession updates an existing session
func (s *SessionService) UpdateSession(sessionID string, chatID *int64) error {
	ctx := context.Background()
	
	// Get current session
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	// Update chat ID
	session.ChatID = chatID

	// Calculate remaining TTL
	ttl := time.Until(*session.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("session expired")
	}

	// Save updated session
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	return s.redis.Set(ctx, s.key(sessionID), data, ttl).Err()
}

// DeleteSession removes a session
func (s *SessionService) DeleteSession(sessionID string) error {
	ctx := context.Background()
	return s.redis.Del(ctx, s.key(sessionID)).Err()
}

// ExtendSession extends the TTL of a session
func (s *SessionService) ExtendSession(sessionID string, duration time.Duration) error {
	ctx := context.Background()
	return s.redis.Expire(ctx, s.key(sessionID), duration).Err()
}

// GetActiveSessions returns count of active sessions
func (s *SessionService) GetActiveSessions() (int64, error) {
	ctx := context.Background()
	keys, err := s.redis.Keys(ctx, "session:*").Result()
	if err != nil {
		return 0, err
	}
	return int64(len(keys)), nil
}

// key generates the Redis key for a session
func (s *SessionService) key(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}