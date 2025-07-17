package services

import (
	"fmt"
	"sync"

	"ai-gateway-hub/internal/config"
	"ai-gateway-hub/internal/models"
	"ai-gateway-hub/internal/providers"
)

// ProviderRegistry manages AI providers
type ProviderRegistry struct {
	providers map[string]providers.AIProvider
	mu        sync.RWMutex
}

func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]providers.AIProvider),
	}
}

// Register adds a provider to the registry
func (r *ProviderRegistry) Register(provider providers.AIProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := provider.GetID()
	if _, exists := r.providers[id]; exists {
		return fmt.Errorf("provider %s already registered", id)
	}

	r.providers[id] = provider
	return nil
}

// Get retrieves a provider by ID
func (r *ProviderRegistry) Get(id string) (providers.AIProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[id]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", id)
	}

	return provider, nil
}

// List returns all registered providers
func (r *ProviderRegistry) List() []*models.Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Provider
	for _, p := range r.providers {
		result = append(result, &models.Provider{
			ID:          p.GetID(),
			Name:        p.GetName(),
			Description: p.GetDescription(),
			Available:   p.IsAvailable(),
		})
	}

	return result
}

// RegisterDefaultProviders registers the default set of providers
func (r *ProviderRegistry) RegisterDefaultProviders(cfg *config.Config) error {
	// Register Claude provider
	claudeProvider := providers.NewClaudeProvider(
		cfg.ClaudeCLIPath,
		cfg.LogDir,
		cfg.ClaudeSkipPermissions,
		cfg.ClaudeExtraArgs,
	)
	if err := r.Register(claudeProvider); err != nil {
		return fmt.Errorf("failed to register Claude provider: %w", err)
	}

	// Future: Register Gemini provider
	// geminiProvider := providers.NewGeminiProvider(cfg.GeminiCLIPath, cfg.LogDir)
	// if err := r.Register(geminiProvider); err != nil {
	//     return fmt.Errorf("failed to register Gemini provider: %w", err)
	// }

	return nil
}