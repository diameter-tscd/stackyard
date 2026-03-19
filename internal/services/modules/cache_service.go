package modules

import (
	"time"

	"stackyard/config"
	"stackyard/pkg/cache"
	"stackyard/pkg/interfaces"
	"stackyard/pkg/logger"
	"stackyard/pkg/registry"
	"stackyard/pkg/response"

	"github.com/labstack/echo/v4"
)

type CacheService struct {
	enabled bool
	store   *cache.Cache[string]
}

func NewCacheService(enabled bool) *CacheService {
	return &CacheService{
		enabled: enabled,
		store:   cache.New[string](),
	}
}

func (s *CacheService) Name() string        { return "Cache Service" }
func (s *CacheService) WireName() string    { return "cache-service" }
func (s *CacheService) Enabled() bool       { return s.enabled }
func (s *CacheService) Get() interface{}    { return s }
func (s *CacheService) Endpoints() []string { return []string{"/cache"} }

type CacheRequest struct {
	Value string `json:"value"`
	TTL   int    `json:"ttl_seconds"` // Optional
}

func (s *CacheService) RegisterRoutes(g *echo.Group) {
	sub := g.Group("/cache")

	// GET /cache/:key
	sub.GET("/:key", func(c echo.Context) error {
		key := c.Param("key")
		val, found := s.store.Get(key)
		if !found {
			return response.NotFound(c, "Key not found or expired")
		}
		return response.Success(c, map[string]string{"key": key, "value": val})
	})

	// POST /cache/:key
	sub.POST("/:key", func(c echo.Context) error {
		key := c.Param("key")
		var req CacheRequest
		if err := c.Bind(&req); err != nil {
			return response.BadRequest(c, "Invalid body")
		}

		ttl := time.Duration(req.TTL) * time.Second
		s.store.Set(key, req.Value, ttl)

		return response.Success(c, map[string]string{
			"message": "Cached successfully",
			"key":     key,
			"ttl":     ttl.String(),
		})
	})
}

// Auto-registration function - called when package is imported
func init() {
	registry.RegisterService("cache_service", func(config *config.Config, logger *logger.Logger, deps *registry.Dependencies) interfaces.Service {
		return NewCacheService(config.Services.IsEnabled("cache_service"))
	})
}
