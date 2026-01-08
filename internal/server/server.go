package server

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"

	"test-go/config"
	"test-go/internal/middleware"
	"test-go/internal/monitoring"
	"test-go/internal/services"
	"test-go/pkg/infrastructure"
	"test-go/pkg/logger"
	"test-go/pkg/response"
	"test-go/pkg/utils"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

type Server struct {
	echo                      *echo.Echo
	config                    *config.Config
	logger                    *logger.Logger
	redisManager              *infrastructure.RedisManager
	kafkaManager              *infrastructure.KafkaManager
	postgresManager           *infrastructure.PostgresManager
	postgresConnectionManager *infrastructure.PostgresConnectionManager
	mongoManager              *infrastructure.MongoManager
	mongoConnectionManager    *infrastructure.MongoConnectionManager
	grafanaManager            *infrastructure.GrafanaManager
	cronManager               *infrastructure.CronManager
	broadcaster               *monitoring.LogBroadcaster
	infraInitManager          *infrastructure.InfraInitManager
}

func New(cfg *config.Config, l *logger.Logger, b *monitoring.LogBroadcaster) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Enable GZIP compression for all responses
	e.Use(echoMiddleware.Gzip())

	// Custom HTTP Error Handler for JSON responses
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		l.Error("HTTP Error", err)

		// Handle HTTP errors with JSON response
		if he, ok := err.(*echo.HTTPError); ok {
			var message string
			code := he.Code

			// Custom message for 404 Not Found
			if code == 404 {
				message = "Endpoint not found. This incident will be reported."
				response.Error(c, code, "ENDPOINT_NOT_FOUND", message, map[string]interface{}{
					"path":   c.Request().URL.Path,
					"method": c.Request().Method,
				})
				return
			}

			// For other HTTP errors, use the original message if it's a string
			if msg, ok := he.Message.(string); ok {
				message = msg
			} else {
				message = "An unexpected error occurred"
			}
			response.Error(c, code, "HTTP_ERROR", message)
			return
		}

		// For non-HTTP errors, return internal server error
		response.InternalServerError(c, "An unexpected error occurred")
	}

	return &Server{
		echo:        e,
		config:      cfg,
		logger:      l,
		broadcaster: b,
	}
}

func (s *Server) Start() error {
	// Initialize async infrastructure manager
	s.infraInitManager = infrastructure.NewInfraInitManager(s.logger)

	// 1. Start Async Infrastructure Initialization (doesn't block)
	s.logger.Info("Starting async infrastructure initialization...")
	s.redisManager, s.kafkaManager, _, s.postgresConnectionManager, s.mongoConnectionManager, s.grafanaManager, s.cronManager =
		s.infraInitManager.StartAsyncInitialization(s.config, s.logger)

	// Set default connections for backward compatibility
	if s.postgresConnectionManager != nil {
		if defaultConn, exists := s.postgresConnectionManager.GetDefaultConnection(); exists {
			s.postgresManager = defaultConn
		}
	}
	if s.mongoConnectionManager != nil {
		if defaultConn, exists := s.mongoConnectionManager.GetDefaultConnection(); exists {
			s.mongoManager = defaultConn
		}
	}

	// 2. Init Middleware (synchronous, lightweight)
	s.logger.Info("Initializing Middleware...")
	middleware.InitMiddlewares(s.echo, middleware.Config{
		AuthType: s.config.Auth.Type,
		Logger:   s.logger,
	})

	// Add encryption middleware if enabled
	if s.config.Encryption.Enabled {
		s.logger.Info("Initializing Encryption Middleware...")
		s.echo.Use(middleware.EncryptionMiddleware(s.config, s.logger))
	}

	// 3. Init Services (phased: independent first, then infrastructure-dependent)
	s.logger.Info("Booting Services...")
	registry := services.NewRegistry(s.logger)

	// Health Check Endpoint with infrastructure status
	s.echo.GET("/health", func(c echo.Context) error {
		health := map[string]interface{}{
			"status":                  "ok",
			"server_ready":            true,
			"infrastructure":          s.infraInitManager.GetStatus(),
			"initialization_progress": s.infraInitManager.GetInitializationProgress(),
		}
		return response.Success(c, health)
	})

	// Infrastructure status endpoint
	s.echo.GET("/health/infrastructure", func(c echo.Context) error {
		status := s.infraInitManager.GetStatus()
		return response.Success(c, status)
	})

	// Restart Endpoint (Maintenance)
	s.echo.POST("/restart", func(c echo.Context) error {
		go func() {
			time.Sleep(500 * time.Millisecond)
			os.Exit(1)
		}()
		return response.Success(c, map[string]string{"status": "restarting", "message": "Service is restarting..."})
	})

	// Create service registrar and register all services
	serviceRegistrar := services.NewServiceRegistrar(
		s.config,
		s.logger,
		s.redisManager,
		s.kafkaManager,
		s.postgresManager,
		s.postgresConnectionManager,
		s.mongoManager,
		s.mongoConnectionManager,
		s.grafanaManager,
		s.cronManager,
	)

	// Register all services (simple and straightforward)
	serviceRegistrar.RegisterAllServices(registry, s.echo)
	s.logger.Info("All services registered successfully, ready to start monitoring")

	// 4. Start Monitoring (if enabled) - after all services are registered
	if s.config.Monitoring.Enabled {
		// Dynamic Service List Generation
		var servicesList []monitoring.ServiceInfo
		for _, srv := range registry.GetServices() {
			// Prepend /api/v1 to endpoints
			var fullEndpoints []string
			for _, endp := range srv.Endpoints() {
				fullEndpoints = append(fullEndpoints, "/api/v1"+endp)
			}

			servicesList = append(servicesList, monitoring.ServiceInfo{
				Name:       srv.Name(),
				StructName: reflect.TypeOf(srv).Elem().String(),
				Active:     srv.Enabled(),
				Endpoints:  fullEndpoints,
			})
		}
		go monitoring.Start(s.config.Monitoring, s.config, s, s.broadcaster, s.redisManager, s.postgresManager, s.postgresConnectionManager, s.mongoManager, s.mongoConnectionManager, s.kafkaManager, s.cronManager, servicesList, s.logger)
		s.logger.Info("Monitoring interface started", "port", s.config.Monitoring.Port, "services_count", len(servicesList))
	}

	// 5. Start HTTP Server immediately (doesn't wait for infrastructure)
	port := s.config.Server.Port
	s.logger.Info("HTTP server starting immediately", "port", port, "env", s.config.App.Env)
	s.logger.Info("Infrastructure components initializing in background...")

	return s.echo.Start(":" + port)
}

// GetStatus satisfies monitoring.StatusProvider
func (s *Server) GetStatus() map[string]interface{} {
	diskStats, _ := utils.GetDiskUsage()
	netStats, _ := utils.GetNetworkInfo()

	infra := map[string]bool{
		"redis":    s.config.Redis.Enabled && s.redisManager != nil,
		"kafka":    s.config.Kafka.Enabled && s.kafkaManager != nil,
		"postgres": (s.config.Postgres.Enabled || s.config.PostgresMultiConfig.Enabled) && (s.postgresManager != nil || s.postgresConnectionManager != nil),
		"mongo":    (s.config.Mongo.Enabled || s.config.MongoMultiConfig.Enabled) && (s.mongoManager != nil || s.mongoConnectionManager != nil),
		"grafana":  s.config.Grafana.Enabled && s.grafanaManager != nil,
		"cron":     s.config.Cron.Enabled && s.cronManager != nil,
	}

	return map[string]interface{}{
		"version":        "1.0.0",
		"services":       s.config.Services, // Dynamic map from config
		"infrastructure": infra,
		"system": map[string]interface{}{
			"disk":    diskStats,
			"network": netStats,
		},
	}
}

// Shutdown performs graceful shutdown of all infrastructure components
func (s *Server) Shutdown(ctx context.Context, logger *logger.Logger) error {
	logger.Info("Starting graceful shutdown of infrastructure...")

	// Force shutdown when more 10s
	go func() {
		warnTimeout := "Maximum shutdown time is 20s, force shutdown when timeout."
		warnForce := "Graceful shutdown timed out, force shutdown."
		duration := 10 * time.Second

		if logger != nil {
			logger.Warn(warnTimeout)
			time.Sleep(duration)
			logger.Fatal(warnForce, nil)
		}

		fmt.Println(warnTimeout)
		time.Sleep(duration)
		os.Exit(1)

	}()

	// Stop async initialization manager
	if s.infraInitManager != nil {
		logger.Info("Stopping async infrastructure initialization manager...")
		// Note: InfraInitManager doesn't have a Close method, but we can signal completion
	}

	// Shutdown infrastructure components in reverse order of initialization
	var shutdownErrors []error

	// 1. Cron Manager
	if s.cronManager != nil {
		logger.Info("Shutting down Cron Manager...")
		if err := s.cronManager.Close(); err != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("cron manager shutdown error: %w", err))
			logger.Error("Error shutting down Cron Manager", err)
		} else {
			logger.Info("Cron Manager shut down successfully")
		}
	}

	// 2. MongoDB connections
	if s.mongoConnectionManager != nil {
		logger.Info("Shutting down MongoDB connections...")
		if err := s.mongoConnectionManager.CloseAll(); err != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("mongodb shutdown error: %w", err))
			logger.Error("Error shutting down MongoDB connections", err)
		} else {
			logger.Info("MongoDB connections shut down successfully")
		}
	}

	// 3. PostgreSQL connections
	if s.postgresConnectionManager != nil {
		logger.Info("Shutting down PostgreSQL connections...")
		if err := s.postgresConnectionManager.CloseAll(); err != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("postgres shutdown error: %w", err))
			logger.Error("Error shutting down PostgreSQL connections", err)
		} else {
			logger.Info("PostgreSQL connections shut down successfully")
		}
	}

	// 4. Kafka Manager
	if s.kafkaManager != nil {
		logger.Info("Shutting down Kafka Manager...")
		if err := s.kafkaManager.Close(); err != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("kafka shutdown error: %w", err))
			logger.Error("Error shutting down Kafka Manager", err)
		} else {
			logger.Info("Kafka Manager shut down successfully")
		}
	}

	// 5. Redis Manager
	if s.redisManager != nil {
		logger.Info("Shutting down Redis Manager...")
		if err := s.redisManager.Close(); err != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("redis shutdown error: %w", err))
			logger.Error("Error shutting down Redis Manager", err)
		} else {
			logger.Info("Redis Manager shut down successfully")
		}
	}

	// Log shutdown summary
	if len(shutdownErrors) > 0 {
		logger.Warn("Graceful shutdown completed with errors", "error_count", len(shutdownErrors))
		for _, err := range shutdownErrors {
			logger.Error("Shutdown error", err)
		}
		return fmt.Errorf("shutdown completed with %d errors", len(shutdownErrors))
	} else {
		logger.Info("Graceful shutdown completed successfully")
		return nil
	}
}
