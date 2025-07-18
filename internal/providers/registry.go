package providers

import (
	"fmt"
	"sync"
)

// Registry manages AI providers
type Registry struct {
	providers map[string]AIProvider
	mu        sync.RWMutex
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]AIProvider),
	}
}

// Register adds a new provider to the registry
func (r *Registry) Register(name string, provider AIProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}
	
	r.providers[name] = provider
	return nil
}

// Get retrieves a provider by name
func (r *Registry) Get(name string) (AIProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	provider, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	
	return provider, nil
}

// List returns all registered providers
func (r *Registry) List() []AIProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	providers := make([]AIProvider, 0, len(r.providers))
	for _, p := range r.providers {
		providers = append(providers, p)
	}
	
	return providers
}

// Remove removes a provider from the registry
func (r *Registry) Remove(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	delete(r.providers, name)
}