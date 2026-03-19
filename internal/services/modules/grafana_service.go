package modules

import (
	"stackyard/config"
	"stackyard/pkg/infrastructure"
	"stackyard/pkg/interfaces"
	"stackyard/pkg/logger"
	"stackyard/pkg/registry"
	"stackyard/pkg/response"
	"strconv"

	"github.com/labstack/echo/v4"
)

// ServiceI provides Grafana integration endpoints
type GrafanaService struct {
	grafanaManager *infrastructure.GrafanaManager
	enabled        bool
	logger         *logger.Logger
}

func NewGrafanaService(grafanaManager *infrastructure.GrafanaManager, enabled bool, logger *logger.Logger) *GrafanaService {
	return &GrafanaService{
		grafanaManager: grafanaManager,
		enabled:        enabled,
		logger:         logger,
	}
}

func (s *GrafanaService) Name() string        { return "Grafana Service" }
func (s *GrafanaService) WireName() string    { return "grafana-service" }
func (s *GrafanaService) Enabled() bool       { return s.enabled && s.grafanaManager != nil }
func (s *GrafanaService) Endpoints() []string { return []string{"/grafana"} }
func (s *GrafanaService) Get() interface{}    { return s }

func (s *GrafanaService) RegisterRoutes(g *echo.Group) {
	sub := g.Group("/grafana")

	// Dashboard management
	sub.POST("/dashboards", s.createDashboard)
	sub.PUT("/dashboards/:uid", s.updateDashboard)
	sub.GET("/dashboards/:uid", s.getDashboard)
	sub.DELETE("/dashboards/:uid", s.deleteDashboard)
	sub.GET("/dashboards", s.listDashboards)

	// Data source management
	sub.POST("/datasources", s.createDataSource)

	// Annotations
	sub.POST("/annotations", s.createAnnotation)

	// Health check
	sub.GET("/health", s.getHealth)
}

// createDashboard creates a new Grafana dashboard
func (s *GrafanaService) createDashboard(c echo.Context) error {
	var dashboard infrastructure.GrafanaDashboard
	if err := c.Bind(&dashboard); err != nil {
		return response.BadRequest(c, "Invalid dashboard data")
	}

	result, err := s.grafanaManager.CreateDashboard(c.Request().Context(), dashboard)
	if err != nil {
		s.logger.Error("Failed to create Grafana dashboard", err)
		return response.InternalServerError(c, "Failed to create dashboard")
	}

	return response.Created(c, result, "Dashboard created successfully")
}

// updateDashboard updates an existing Grafana dashboard
func (s *GrafanaService) updateDashboard(c echo.Context) error {
	uid := c.Param("uid")
	if uid == "" {
		return response.BadRequest(c, "Dashboard UID is required")
	}

	var dashboard infrastructure.GrafanaDashboard
	if err := c.Bind(&dashboard); err != nil {
		return response.BadRequest(c, "Invalid dashboard data")
	}

	// Ensure UID is set
	dashboard.UID = uid

	result, err := s.grafanaManager.UpdateDashboard(c.Request().Context(), dashboard)
	if err != nil {
		s.logger.Error("Failed to update Grafana dashboard", err, "uid", uid)
		return response.InternalServerError(c, "Failed to update dashboard")
	}

	return response.Success(c, result, "Dashboard updated successfully")
}

// getDashboard retrieves a Grafana dashboard by UID
func (s *GrafanaService) getDashboard(c echo.Context) error {
	uid := c.Param("uid")
	if uid == "" {
		return response.BadRequest(c, "Dashboard UID is required")
	}

	dashboard, err := s.grafanaManager.GetDashboard(c.Request().Context(), uid)
	if err != nil {
		s.logger.Error("Failed to get Grafana dashboard", err, "uid", uid)
		return response.NotFound(c, "Dashboard not found")
	}

	return response.Success(c, dashboard, "Dashboard retrieved successfully")
}

// deleteDashboard deletes a Grafana dashboard by UID
func (s *GrafanaService) deleteDashboard(c echo.Context) error {
	uid := c.Param("uid")
	if uid == "" {
		return response.BadRequest(c, "Dashboard UID is required")
	}

	err := s.grafanaManager.DeleteDashboard(c.Request().Context(), uid)
	if err != nil {
		s.logger.Error("Failed to delete Grafana dashboard", err, "uid", uid)
		return response.InternalServerError(c, "Failed to delete dashboard")
	}

	return response.Success(c, nil, "Dashboard deleted successfully")
}

// listDashboards lists all Grafana dashboards
func (s *GrafanaService) listDashboards(c echo.Context) error {
	// Parse pagination parameters
	page := 1
	perPage := 50

	if pageStr := c.QueryParam("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if perPageStr := c.QueryParam("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}

	dashboards, err := s.grafanaManager.ListDashboards(c.Request().Context())
	if err != nil {
		s.logger.Error("Failed to list Grafana dashboards", err)
		return response.InternalServerError(c, "Failed to list dashboards")
	}

	// Simple pagination (in a real implementation, you'd use proper pagination)
	start := (page - 1) * perPage
	end := start + perPage

	if start >= len(dashboards) {
		dashboards = []infrastructure.GrafanaDashboard{}
	} else if end > len(dashboards) {
		dashboards = dashboards[start:]
	} else {
		dashboards = dashboards[start:end]
	}

	meta := response.CalculateMeta(page, perPage, int64(len(dashboards)))
	return response.SuccessWithMeta(c, dashboards, meta, "Dashboards retrieved successfully")
}

// createDataSource creates a new Grafana data source
func (s *GrafanaService) createDataSource(c echo.Context) error {
	var ds infrastructure.GrafanaDataSource
	if err := c.Bind(&ds); err != nil {
		return response.BadRequest(c, "Invalid data source data")
	}

	result, err := s.grafanaManager.CreateDataSource(c.Request().Context(), ds)
	if err != nil {
		s.logger.Error("Failed to create Grafana data source", err)
		return response.InternalServerError(c, "Failed to create data source")
	}

	return response.Created(c, result, "Data source created successfully")
}

// createAnnotation creates a new Grafana annotation
func (s *GrafanaService) createAnnotation(c echo.Context) error {
	var annotation infrastructure.GrafanaAnnotation
	if err := c.Bind(&annotation); err != nil {
		return response.BadRequest(c, "Invalid annotation data")
	}

	result, err := s.grafanaManager.CreateAnnotation(c.Request().Context(), annotation)
	if err != nil {
		s.logger.Error("Failed to create Grafana annotation", err)
		return response.InternalServerError(c, "Failed to create annotation")
	}

	return response.Created(c, result, "Annotation created successfully")
}

// getHealth returns Grafana health status
func (s *GrafanaService) getHealth(c echo.Context) error {
	health, err := s.grafanaManager.GetHealth(c.Request().Context())
	if err != nil {
		s.logger.Error("Failed to get Grafana health", err)
		return response.ServiceUnavailable(c, "Grafana is not available")
	}

	return response.Success(c, health, "Grafana health check successful")
}

// Auto-registration function - called when package is imported
func init() {
	registry.RegisterService("grafana_service", func(config *config.Config, logger *logger.Logger, deps *registry.Dependencies) interfaces.Service {
		if !config.Services.IsEnabled("grafana_service") {
			return nil
		}
		if deps == nil || deps.GrafanaManager == nil {
			logger.Warn("Grafana manager not available, skipping Grafana Service")
			return nil
		}
		return NewGrafanaService(deps.GrafanaManager, true, logger)
	})
}
