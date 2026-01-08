package monitoring

import (
	"net/http"
	"test-go/config"
	"test-go/internal/monitoring/database"
	"test-go/internal/monitoring/session"
	"test-go/pkg/infrastructure"
	"test-go/pkg/logger"
	"time"

	monMiddleware "test-go/internal/monitoring/middleware"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type StatusProvider interface {
	GetStatus() map[string]interface{}
}

type ServiceInfo struct {
	Name       string   `json:"name"`
	StructName string   `json:"struct_name"`
	Active     bool     `json:"active"`
	Endpoints  []string `json:"endpoints"`
}

func Start(
	cfg config.MonitoringConfig,
	appConfig *config.Config,
	statusProvider StatusProvider,
	broadcaster *LogBroadcaster,
	redis *infrastructure.RedisManager,
	postgres *infrastructure.PostgresManager,
	postgresConnectionManager *infrastructure.PostgresConnectionManager,
	mongo *infrastructure.MongoManager,
	mongoConnectionManager *infrastructure.MongoConnectionManager,
	kafka *infrastructure.KafkaManager,
	cron *infrastructure.CronManager,
	services []ServiceInfo,
	log *logger.Logger,
) {
	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Warn("Failed to initialize user settings database", "error", err)
	} else {
		log.Info("User settings database initialized")

		// Ensure upload directory exists
		uploadDir := cfg.UploadDir
		if uploadDir == "" {
			uploadDir = "web/monitoring/uploads"
		}
		if err := database.EnsureUploadDirectory(uploadDir); err != nil {
			log.Warn("Failed to create upload directory", "error", err)
		}

		// Create default user if not exists
		settings, _ := database.GetUserSettings()
		if settings == nil {
			if err := database.CreateDefaultUser(cfg.Password); err != nil {
				log.Warn("Failed to create default user", "error", err)
			} else {
				log.Info("Default user created")
			}
		}
	}

	// Initialize Infrastructure Managers
	minioMgr, err := infrastructure.NewMinIOManager(appConfig.Monitoring.MinIO)
	if err != nil {
		log.Error("Failed to connect to MinIO", err)
	} else {
		log.Info("MinIO Manager initialized")
	}

	systemMgr := infrastructure.NewSystemManager()
	httpMgr := infrastructure.NewHttpManager(appConfig.Monitoring.External)

	// Initialize session manager
	sessionManager := session.NewManager(24 * time.Hour)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip()) // Enable GZIP compression for all responses
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  []string{"*"},
		AllowHeaders:  []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, "X-Correlation-ID"},
		ExposeHeaders: []string{"X-Obfuscated"},
	}))
	e.Use(monMiddleware.Obfuscator(cfg.ObfuscateAPI))

	// Public routes (no auth required)
	e.GET("/", func(c echo.Context) error {
		return c.File("web/monitoring/login.html")
	})
	e.Static("/assets", "web/monitoring/assets")

	// Auth endpoints
	e.POST("/login", handleLogin(sessionManager))
	e.POST("/logout", handleLogout(sessionManager))

	// Protected routes group (require session)
	protected := e.Group("")
	protected.Use(session.Middleware(sessionManager))

	// Dashboard and API routes (protected)
	protected.GET("/dashboard", func(c echo.Context) error {
		return c.File("web/monitoring/index.html")
	})
	protected.Static("/api/user/photos", appConfig.Monitoring.UploadDir+"/profiles")

	// Register API Handlers
	h := &Handler{
		config:                    appConfig,
		statusProvider:            statusProvider,
		broadcaster:               broadcaster,
		redis:                     redis,
		postgres:                  postgres,
		postgresConnectionManager: postgresConnectionManager,
		mongo:                     mongo,
		mongoConnectionManager:    mongoConnectionManager,
		kafka:                     kafka,
		cron:                      cron,
		services:                  services,
		minio:                     minioMgr,
		system:                    systemMgr,
		http:                      httpMgr,
	}
	h.RegisterRoutes(protected)

	log.Info("Monitoring UI running", "url", "http://localhost:"+cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
		log.Error("Failed to start monitoring server", err)
	}
}
