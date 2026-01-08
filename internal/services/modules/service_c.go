package modules

import (
	"time"

	"test-go/pkg/cache"
	"test-go/pkg/response"

	"github.com/labstack/echo/v4"
)

type ServiceC struct {
	enabled bool
	store   *cache.Cache[string]
}

func NewServiceC(enabled bool) *ServiceC {
	return &ServiceC{
		enabled: enabled,
		store:   cache.New[string](),
	}
}

func (s *ServiceC) Name() string        { return "Service C (Cache Demo)" }
func (s *ServiceC) Enabled() bool       { return s.enabled }
func (s *ServiceC) Endpoints() []string { return []string{"/cache"} }

type CacheRequest struct {
	Value string `json:"value"`
	TTL   int    `json:"ttl_seconds"` // Optional
}

func (s *ServiceC) RegisterRoutes(g *echo.Group) {
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
