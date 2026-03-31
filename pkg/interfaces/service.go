package interfaces

import (
	"github.com/labstack/echo/v4"
)

// Service defines the interface that all services must implement
type Service interface {
	// Name returns the human-readable name of the service
	Name() string

	// Alias Name for dependency injection
	WireName() string

	// Enabled returns whether the service is enabled
	Enabled() bool

	// Endpoints returns a list of endpoint patterns this service handles
	Endpoints() []string

	// RegisterRoutes registers the service's routes with the Echo router
	RegisterRoutes(g *echo.Group)

	// Get service
	Get() interface{}
}
