package modules

import (
	"stackyard/config"
	"stackyard/pkg/registry"
	"stackyard/pkg/interfaces"
	"stackyard/pkg/logger"
	"stackyard/pkg/response"

	"github.com/labstack/echo/v4"
)

type ServiceB struct {
	enabled bool
}

func NewServiceB(enabled bool) *ServiceB {
	return &ServiceB{enabled: enabled}
}

func (s *ServiceB) Name() string        { return "Service B (Products)" }
func (s *ServiceB) Enabled() bool       { return s.enabled }
func (s *ServiceB) Endpoints() []string { return []string{"/products"} }

func (s *ServiceB) RegisterRoutes(g *echo.Group) {
	sub := g.Group("/products")
	sub.GET("", func(c echo.Context) error {
		return response.Success(c, map[string]string{"message": "Hello from Service B - Products"})
	})
}

// Auto-registration function - called when package is imported
func init() {
	registry.RegisterService("service_b", func(config *config.Config, logger *logger.Logger, deps *registry.Dependencies) interfaces.Service {
		return NewServiceB(config.Services.IsEnabled("service_b"))
	})
}
