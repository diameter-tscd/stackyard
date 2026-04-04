package modules

import (
	"stackyrd/config"
	"stackyrd/pkg/interfaces"
	"stackyrd/pkg/logger"
	"stackyrd/pkg/registry"
	"stackyrd/pkg/response"

	"github.com/labstack/echo/v4"
	"stackyrd/pkg/infrastructure"
)

type Papp struct {
	enabled bool
	postgresManager *infrastructure.PostgresManager
	mongoConnectionManager *infrastructure.MongoConnectionManager
	postgresConnectionManager *infrastructure.PostgresConnectionManager
	logger *logger.Logger
}

func NewPapp(
	enabled bool,
	postgresManager *infrastructure.PostgresManager,
	mongoConnectionManager *infrastructure.MongoConnectionManager,
	postgresConnectionManager *infrastructure.PostgresConnectionManager,
	logger *logger.Logger,
) *Papp {
	return &Papp{
		enabled: enabled,
		postgresManager: postgresManager,
		mongoConnectionManager: mongoConnectionManager,
		postgresConnectionManager: postgresConnectionManager,
		logger: logger,
	}
}

func (s *Papp) Name() string     { return "Papp Service" }
func (s *Papp) WireName() string { return "papp-service" }
func (s *Papp) Enabled() bool    { return s.enabled }
func (s *Papp) Get() interface{} { return s }
func (s *Papp) Endpoints() []string {
	return []string{"/papp"}
}

// RegisterRoutes registers the service routes
func (s *Papp) RegisterRoutes(g *echo.Group) {
	sub := g.Group("/papp")
	sub.GET("", s.listHandler)
	sub.POST("", s.createHandler)
	sub.GET("/:id", s.getHandler)
	sub.PUT("/:id", s.updateHandler)
	sub.DELETE("/:id", s.deleteHandler)
}

// Handler methods (implement your business logic here)

func (s *Papp) listHandler(c echo.Context) error {
	// TODO: Implement list logic
	return response.Success(c, []interface{}{}, "List endpoint")
}

func (s *Papp) createHandler(c echo.Context) error {
	// TODO: Implement create logic
	return response.Created(c, nil, "Create endpoint")
}

func (s *Papp) getHandler(c echo.Context) error {
	// TODO: Implement get logic
	id := c.Param("id")
	return response.Success(c, map[string]string{"id": id}, "Get endpoint")
}

func (s *Papp) updateHandler(c echo.Context) error {
	// TODO: Implement update logic
	id := c.Param("id")
	return response.Success(c, map[string]string{"id": id}, "Update endpoint")
}

func (s *Papp) deleteHandler(c echo.Context) error {
	// TODO: Implement delete logic
	return response.NoContent(c)
}

// Auto-registration function - called when package is imported
func init() {
	registry.RegisterService("papp_service", func(config *config.Config, logger *logger.Logger, deps *registry.Dependencies) interfaces.Service {
		helper := registry.NewServiceHelper(config, logger, deps)
		
		if !helper.IsServiceEnabled("papp_service") {
			return nil
		}
		
		postgresManager, ok := registry.GetTyped[*infrastructure.PostgresManager](deps, "postgresManager")
		if !ok {
			logger.Warn("PostgresManager not available, skipping service")
			return nil
		}
		
		mongoConnectionManager, ok := registry.GetTyped[*infrastructure.MongoConnectionManager](deps, "mongoConnectionManager")
		if !ok {
			logger.Warn("MongoConnectionManager not available, skipping service")
			return nil
		}
		
		postgresConnectionManager, ok := registry.GetTyped[*infrastructure.PostgresConnectionManager](deps, "postgresConnectionManager")
		if !ok {
			logger.Warn("PostgresConnectionManager not available, skipping service")
			return nil
		}
		
		return NewPapp(true, postgresManager, mongoConnectionManager, postgresConnectionManager, logger)
	})
}