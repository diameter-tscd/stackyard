package infrastructure

import (
	"fmt"
	"stackyrd/config"
	"stackyrd/pkg/logger"
	"sync"
	"time"
)

// ComponentRegistry manages all infrastructure components
type ComponentRegistry struct {
	components       sync.Map // map[string]InfrastructureComponent — write-once after boot
	factories        sync.Map // map[string]ComponentFactory
	cachedComponents sync.Map // map[string]map[string]InfrastructureComponent — TTL-cached GetAll snapshot
	cacheExpiry      time.Time
	cacheMu          sync.Mutex
	cacheTTL         time.Duration
}

// Global registry instance
var (
	globalRegistry *ComponentRegistry
	registryOnce   sync.Once
)

// GetGlobalRegistry returns the singleton registry instance
func GetGlobalRegistry() *ComponentRegistry {
	registryOnce.Do(func() {
		globalRegistry = &ComponentRegistry{
			cacheTTL: 500 * time.Millisecond,
		}
	})
	return globalRegistry
}

// RegisterComponent registers a component factory with the global registry
func RegisterComponent(name string, factory ComponentFactory) {
	GetGlobalRegistry().Register(name, factory)
}

// Register registers a component factory
func (r *ComponentRegistry) Register(name string, factory ComponentFactory) {
	r.factories.Store(name, factory)
}

// Initialize initializes all registered components
func (r *ComponentRegistry) Initialize(cfg *config.Config, logger *logger.Logger) error {
	r.cacheMu.Lock()
	defer r.cacheMu.Unlock()

	r.factories.Range(func(nameObj, factoryObj interface{}) bool {
		name := nameObj.(string)
		factory := factoryObj.(ComponentFactory)
		component, err := factory(cfg, logger)
		if err != nil {
			logger.Error("Failed to initialize "+name, err)
			return true
		}
		if component != nil {
			r.components.Store(name, component)
			logger.Info(name + " initialized")
		}
		return true
	})
	return nil
}

// Get retrieves a component by name — lock-free read path via sync.Map
func (r *ComponentRegistry) Get(name string) (InfrastructureComponent, bool) {
	component, ok := r.components.Load(name)
	if !ok {
		return nil, false
	}
	return component.(InfrastructureComponent), true
}

// GetAll returns all components — returns a TTL-cached snapshot to avoid
// re-allocating and copying the entire map on every /health/dependencies call.
func (r *ComponentRegistry) GetAll() map[string]InfrastructureComponent {
	// Fast path: return cached snapshot when still within TTL
	r.cacheMu.Lock()
	if time.Now().Before(r.cacheExpiry) {
		if cached, ok := r.cachedComponents.Load("__all__"); ok {
			result := cached.(map[string]InfrastructureComponent)
			r.cacheMu.Unlock()
			return result
		}
	}
	r.cacheMu.Unlock()

	// Slow path: rebuild snapshot under cacheMu
	r.cacheMu.Lock()
	defer r.cacheMu.Unlock()
	if time.Now().Before(r.cacheExpiry) {
		if cached, ok := r.cachedComponents.Load("__all__"); ok {
			return cached.(map[string]InfrastructureComponent)
		}
	}
	result := make(map[string]InfrastructureComponent)
	r.components.Range(func(key, value interface{}) bool {
		result[key.(string)] = value.(InfrastructureComponent)
		return true
	})
	r.cachedComponents.Store("__all__", result)
	r.cacheExpiry = time.Now().Add(r.cacheTTL)
	return result
}

// CloseAll closes all components
func (r *ComponentRegistry) CloseAll() []error {
	var errors []error
	r.components.Range(func(key, value interface{}) bool {
		if err := value.(InfrastructureComponent).Close(); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", key.(string), err))
		}
		return true
	})
	return errors
}
