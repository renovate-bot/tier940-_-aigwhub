package providers

import (
	"context"
	"io"
)

// AIProvider defines the interface for AI providers
type AIProvider interface {
	// GetID returns the unique identifier for this provider
	GetID() string

	// GetName returns the display name of this provider
	GetName() string

	// GetDescription returns a brief description of this provider
	GetDescription() string

	// IsAvailable checks if the provider is available and configured
	IsAvailable() bool

	// SendPrompt sends a prompt to the AI and returns a response reader
	SendPrompt(ctx context.Context, prompt string, chatID int64) (io.ReadCloser, error)

	// StreamResponse streams the response to the provided writer
	StreamResponse(ctx context.Context, prompt string, chatID int64, writer io.Writer) error
}