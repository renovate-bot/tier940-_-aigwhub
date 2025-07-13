package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Note: Environment variable helpers removed - use viper for configuration management

// Context helpers
func NewContextWithTimeout(duration time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), duration)
}

func NewContext() context.Context {
	return context.Background()
}

// JSON helpers
func MarshalJSON(v interface{}) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return data, nil
}

func UnmarshalJSON(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}

// Error helpers
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

func NewError(message string, args ...interface{}) error {
	return fmt.Errorf(message, args...)
}

// Mutex helpers
type SafeOperation func() error

func WithLock(mu *sync.Mutex, op SafeOperation) error {
	mu.Lock()
	defer mu.Unlock()
	return op()
}

func WithRLock(mu *sync.RWMutex, op SafeOperation) error {
	mu.RLock()
	defer mu.RUnlock()
	return op()
}

// File operation helpers
func CreateFile(path string) (*os.File, error) {
	if err := EnsureDirForFile(path); err != nil {
		return nil, err
	}
	
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create/open file %s: %w", path, err)
	}
	return file, nil
}

func WriteToFile(path string, data []byte) error {
	if err := EnsureDirForFile(path); err != nil {
		return err
	}
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	return nil
}

func ReadFromFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return data, nil
}