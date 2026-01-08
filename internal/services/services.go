package services

import (
	"test-go/pkg/logger"

	"github.com/labstack/echo/v4"
)

// Service defines a module that can register routes
type Service interface {
	Name() string
	RegisterRoutes(g *echo.Group)
	Enabled() bool
	Endpoints() []string
}

// Registry holds available services
type Registry struct {
	services []Service
	logger   *logger.Logger
}

// NewRegistry creates a new service registry
func NewRegistry(l *logger.Logger) *Registry {
	return &Registry{
		services: make([]Service, 0),
		logger:   l,
	}
}

// Register adds a service to the registry
func (r *Registry) Register(s Service) {
	r.services = append(r.services, s)
}

// GetServices returns the list of registered services
func (r *Registry) GetServices() []Service {
	return r.services
}

// Boot initializes enabled services and registers their routes
func (r *Registry) Boot(e *echo.Echo) {
	api := e.Group("/api/v1")

	for _, s := range r.services {
		if s.Enabled() {
			r.logger.Info("Starting Service...", "service", s.Name())
			s.RegisterRoutes(api)
			r.logger.Info("Service Started", "service", s.Name())
		} else {
			r.logger.Warn("Service Skipped (Disabled via config)", "service", s.Name())
		}
	}
}

// BootService boots a single service (for dynamic registration)
func (r *Registry) BootService(e *echo.Echo, s Service) {
	if s.Enabled() {
		api := e.Group("/api/v1")
		r.logger.Info("Starting Service...", "service", s.Name())
		s.RegisterRoutes(api)
		r.logger.Info("Service Started", "service", s.Name())
	} else {
		r.logger.Warn("Service Skipped (Disabled via config)", "service", s.Name())
	}
}
