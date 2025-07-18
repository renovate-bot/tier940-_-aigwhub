package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"ai-gateway-hub/internal/config"
	"ai-gateway-hub/internal/models"
	"ai-gateway-hub/internal/providers"
	"github.com/go-redis/redis/v8"
)

// ProviderRegistry manages AI providers with Redis-based caching
type ProviderRegistry struct {
	providers   map[string]providers.AIProvider
	mu          sync.RWMutex
	redisClient *redis.Client
	ctx         context.Context
}

func NewProviderRegistry(redisClient *redis.Client) *ProviderRegistry {
	registry := &ProviderRegistry{
		providers:   make(map[string]providers.AIProvider),
		redisClient: redisClient,
		ctx:         context.Background(),
	}
	
	// Start background status update routine
	go registry.backgroundStatusUpdater()
	
	return registry
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

// List returns all registered providers with cached status
func (r *ProviderRegistry) List() []*models.Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Provider
	for _, p := range r.providers {
		provider := &models.Provider{
			ID:          p.GetID(),
			Name:        p.GetName(),
			Description: p.GetDescription(),
		}
		
		// Try to get cached status first
		if cachedStatus := r.getCachedStatus(p.GetID()); cachedStatus != nil {
			provider.Available = cachedStatus.Available
			provider.Status = cachedStatus.Status
			provider.Version = cachedStatus.Version
			provider.Details = cachedStatus.Details
		} else {
			// Fallback to direct status check and cache it
			status := p.GetStatus()
			provider.Available = status.Available
			provider.Status = status.Status
			provider.Version = status.Version
			provider.Details = status.Details
			
			// Cache the status asynchronously
			go r.cacheStatus(p.GetID(), status)
		}
		
		result = append(result, provider)
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

// getCachedStatus retrieves provider status from Redis cache
func (r *ProviderRegistry) getCachedStatus(providerID string) *providers.ProviderStatus {
	if r.redisClient == nil {
		return nil
	}
	
	key := fmt.Sprintf("provider_status:%s", providerID)
	data, err := r.redisClient.Get(r.ctx, key).Result()
	if err != nil {
		return nil
	}
	
	var status providers.ProviderStatus
	if err := json.Unmarshal([]byte(data), &status); err != nil {
		return nil
	}
	
	return &status
}

// cacheStatus stores provider status in Redis cache
func (r *ProviderRegistry) cacheStatus(providerID string, status providers.ProviderStatus) {
	if r.redisClient == nil {
		return
	}
	
	key := fmt.Sprintf("provider_status:%s", providerID)
	data, err := json.Marshal(status)
	if err != nil {
		return
	}
	
	// Cache for 5 minutes
	r.redisClient.Set(r.ctx, key, data, 5*time.Minute)
}

// GetProviderStatus returns cached status for a specific provider
func (r *ProviderRegistry) GetProviderStatus(providerID string) (*providers.ProviderStatus, error) {
	r.mu.RLock()
	provider, exists := r.providers[providerID]
	r.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("provider %s not found", providerID)
	}
	
	// Try cache first
	if cachedStatus := r.getCachedStatus(providerID); cachedStatus != nil {
		return cachedStatus, nil
	}
	
	// Get fresh status and cache it
	status := provider.GetStatus()
	go r.cacheStatus(providerID, status)
	
	return &status, nil
}

// backgroundStatusUpdater periodically updates provider status in cache
func (r *ProviderRegistry) backgroundStatusUpdater() {
	ticker := time.NewTicker(2 * time.Minute) // Update every 2 minutes
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			r.updateAllProviderStatus()
		case <-r.ctx.Done():
			return
		}
	}
}

// updateAllProviderStatus updates status for all providers in background
func (r *ProviderRegistry) updateAllProviderStatus() {
	r.mu.RLock()
	providerMap := make(map[string]providers.AIProvider)
	for id, provider := range r.providers {
		providerMap[id] = provider
	}
	r.mu.RUnlock()
	
	// Update status for each provider concurrently
	for id, provider := range providerMap {
		go func(providerID string, p providers.AIProvider) {
			status := p.GetStatus()
			r.cacheStatus(providerID, status)
		}(id, provider)
	}
}