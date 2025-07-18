package providers

import (
	"context"
	"io"
)

// ProviderStatus represents the detailed status of an AI provider
type ProviderStatus struct {
	Available bool   `json:"available"`
	Status    string `json:"status"` // "ready", "not_installed", "not_configured", "error"
	Version   string `json:"version,omitempty"`
	Details   string `json:"details,omitempty"`
}

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

	// GetStatus returns detailed status information about the provider
	GetStatus() ProviderStatus

	// SendPrompt sends a prompt to the AI and returns a response reader
	SendPrompt(ctx context.Context, prompt string, chatID int64) (io.ReadCloser, error)

	// StreamResponse streams the response to the provided writer
	StreamResponse(ctx context.Context, prompt string, chatID int64, writer io.Writer) error
}