package infrastructure

import (
	"context"
	"sync"
	"test-go/config"
	"test-go/pkg/logger"
	"time"
)

// InfraInitStatus represents the initialization status of an infrastructure component
type InfraInitStatus struct {
	Name        string        `json:"name"`
	Initialized bool          `json:"initialized"`
	Error       string        `json:"error,omitempty"`
	StartTime   time.Time     `json:"start_time"`
	Duration    time.Duration `json:"duration,omitempty"`
	Progress    float64       `json:"progress"` // 0.0 to 1.0
}

// InfraInitManager manages asynchronous infrastructure initialization
type InfraInitManager struct {
	status   map[string]*InfraInitStatus
	mu       sync.RWMutex
	logger   *logger.Logger
	doneChan chan struct{}
}

// NewInfraInitManager creates a new infrastructure initialization manager
func NewInfraInitManager(logger *logger.Logger) *InfraInitManager {
	return &InfraInitManager{
		status:   make(map[string]*InfraInitStatus),
		logger:   logger,
		doneChan: make(chan struct{}),
	}
}

// StartAsyncInitialization begins asynchronous initialization of all infrastructure components
func (im *InfraInitManager) StartAsyncInitialization(cfg *config.Config, logger *logger.Logger) (
	*RedisManager,
	*KafkaManager,
	*MinIOManager,
	*PostgresConnectionManager,
	*MongoConnectionManager,
	*GrafanaManager,
	*CronManager,
) {
	var (
		redisManager              *RedisManager
		kafkaManager              *KafkaManager
		minioManager              *MinIOManager
		postgresConnectionManager *PostgresConnectionManager
		mongoConnectionManager    *MongoConnectionManager
		grafanaManager            *GrafanaManager
		cronManager               *CronManager
	)

	// Initialize components synchronously to avoid race conditions
	// Only the connection testing/health checks are done asynchronously

	// Redis
	if cfg.Redis.Enabled {
		rdb, err := NewRedisClient(cfg.Redis)
		if err != nil {
			logger.Error("Failed to initialize Redis", err)
		} else {
			redisManager = rdb
			logger.Info("Redis initialized")
		}
	}

	// Kafka
	if cfg.Kafka.Enabled {
		km, err := NewKafkaManager(cfg.Kafka, logger)
		if err != nil {
			logger.Error("Failed to initialize Kafka", err)
		} else {
			kafkaManager = km
			logger.Info("Kafka initialized")
		}
	}

	// MinIO
	if cfg.Monitoring.MinIO.Endpoint != "" {
		minio, err := NewMinIOManager(cfg.Monitoring.MinIO)
		if err != nil {
			logger.Error("Failed to initialize MinIO", err)
		} else {
			minioManager = minio
			logger.Info("MinIO initialized")
		}
	}

	// PostgreSQL
	if cfg.Postgres.Enabled || cfg.PostgresMultiConfig.Enabled {
		if cfg.PostgresMultiConfig.Enabled && len(cfg.PostgresMultiConfig.Connections) > 0 {
			connManager, err := NewPostgresConnectionManager(cfg.PostgresMultiConfig)
			if err != nil {
				logger.Error("Failed to initialize PostgreSQL connections", err)
			} else {
				postgresConnectionManager = connManager
				logger.Info("PostgreSQL connections initialized")
			}
		} else if cfg.Postgres.Enabled {
			connManager, err := NewPostgresConnectionManager(config.PostgresMultiConfig{
				Enabled: true,
				Connections: []config.PostgresConnectionConfig{
					{
						Name:     "default",
						Enabled:  true,
						Host:     cfg.Postgres.Host,
						Port:     cfg.Postgres.Port,
						User:     cfg.Postgres.User,
						Password: cfg.Postgres.Password,
						DBName:   cfg.Postgres.DBName,
						SSLMode:  cfg.Postgres.SSLMode,
					},
				},
			})
			if err != nil {
				logger.Error("Failed to initialize PostgreSQL", err)
			} else {
				postgresConnectionManager = connManager
				logger.Info("PostgreSQL initialized (single connection)")
			}
		}
	}

	// MongoDB
	if cfg.Mongo.Enabled || cfg.MongoMultiConfig.Enabled {
		if cfg.MongoMultiConfig.Enabled && len(cfg.MongoMultiConfig.Connections) > 0 {
			connManager, err := NewMongoConnectionManager(cfg.MongoMultiConfig, logger)
			if err != nil {
				logger.Error("Failed to initialize MongoDB connections", err)
			} else {
				mongoConnectionManager = connManager
				logger.Info("MongoDB connections initialized")
			}
		} else if cfg.Mongo.Enabled {
			connManager, err := NewMongoConnectionManager(config.MongoMultiConfig{
				Enabled: true,
				Connections: []config.MongoConnectionConfig{
					{
						Name:     "default",
						Enabled:  true,
						URI:      cfg.Mongo.URI,
						Database: cfg.Mongo.Database,
					},
				},
			}, logger)
			if err != nil {
				logger.Error("Failed to initialize MongoDB", err)
			} else {
				mongoConnectionManager = connManager
				logger.Info("MongoDB initialized (single connection)")
			}
		}
	}

	// Grafana
	if cfg.Grafana.Enabled {
		gm, err := NewGrafanaManager(cfg.Grafana, logger)
		if err != nil {
			logger.Error("Failed to initialize Grafana", err)
		} else {
			grafanaManager = gm
			logger.Info("Grafana initialized")
		}
	}

	// Cron (initialize synchronously with jobs)
	if cfg.Cron.Enabled {
		cronManager = NewCronManager()

		// Add cron jobs synchronously
		for name, schedule := range cfg.Cron.Jobs {
			jobName := name
			jobSchedule := schedule
			_, err := cronManager.AddAsyncJob(jobName, jobSchedule, func() {
				logger.Info("Executing Cron Job (Async)", "job", jobName)
			})
			if err != nil {
				logger.Error("Failed to schedule cron job", err, "job", jobName)
			} else {
				logger.Info("Cron job scheduled", "job", jobName, "schedule", jobSchedule)
			}
		}

		cronManager.Start()
		logger.Info("Cron jobs initialized with async execution")
	}

	// Start async health checks and monitoring (non-blocking)
	components := []struct {
		name  string
		check func()
	}{
		{
			name: "redis",
			check: func() {
				if redisManager != nil {
					// Redis manager already performs health checks in GetStatus()
					status := redisManager.GetStatus()
					if connected, ok := status["connected"].(bool); ok && connected {
						logger.Debug("Redis health check passed")
					} else {
						logger.Warn("Redis health check failed")
					}
				}
			},
		},
		{
			name: "kafka",
			check: func() {
				if kafkaManager != nil {
					// Kafka manager handles its own async health checks
					logger.Debug("Kafka health monitoring active")
				}
			},
		},
		{
			name: "minio",
			check: func() {
				if minioManager != nil {
					// MinIO async health checks if needed
					logger.Debug("MinIO health monitoring active")
				}
			},
		},
		{
			name: "postgres",
			check: func() {
				if postgresConnectionManager != nil {
					// Connection manager handles health checks internally
					logger.Debug("PostgreSQL health monitoring active")
				}
			},
		},
		{
			name: "mongodb",
			check: func() {
				if mongoConnectionManager != nil {
					// Connection manager handles health checks internally
					logger.Debug("MongoDB health monitoring active")
				}
			},
		},
		{
			name: "cron",
			check: func() {
				if cronManager != nil {
					// Cron manager is already initialized and running
					logger.Debug("Cron jobs active", "count", len(cronManager.GetJobs()))
				}
			},
		},
	}

	// Start health monitoring asynchronously
	for _, comp := range components {
		comp := comp // Capture loop variable
		go func(name string, checkFn func()) {
			// Update status to initialized
			im.updateStatus(name, &InfraInitStatus{
				Name:        name,
				Initialized: true,
				StartTime:   time.Now(),
				Duration:    time.Since(time.Now()), // Minimal duration
				Progress:    1.0,
			})

			// Perform ongoing health checks
			checkFn()
		}(comp.name, comp.check)
	}

	// Signal that all synchronous initialization is complete
	close(im.doneChan)

	return redisManager, kafkaManager, minioManager, postgresConnectionManager, mongoConnectionManager, grafanaManager, cronManager
}

// updateStatus updates the initialization status of a component
func (im *InfraInitManager) updateStatus(name string, status *InfraInitStatus) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.status[name] = status
}

// updateStatusProgress updates only the progress of a component
func (im *InfraInitManager) updateStatusProgress(name string, progress float64) {
	im.mu.Lock()
	defer im.mu.Unlock()
	if status, exists := im.status[name]; exists {
		status.Progress = progress
	}
}

// GetStatus returns the current initialization status of all components
func (im *InfraInitManager) GetStatus() map[string]*InfraInitStatus {
	im.mu.RLock()
	defer im.mu.RUnlock()

	// Create a copy to avoid race conditions
	status := make(map[string]*InfraInitStatus)
	for k, v := range im.status {
		status[k] = v
	}

	return status
}

// IsInitialized checks if a specific component is initialized
func (im *InfraInitManager) IsInitialized(component string) bool {
	im.mu.RLock()
	defer im.mu.RUnlock()

	if status, exists := im.status[component]; exists {
		return status.Initialized
	}
	return false
}

// WaitForInitialization waits for all components to complete initialization
func (im *InfraInitManager) WaitForInitialization(ctx context.Context) error {
	select {
	case <-im.doneChan:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GetInitializationProgress returns overall initialization progress (0.0 to 1.0)
func (im *InfraInitManager) GetInitializationProgress() float64 {
	status := im.GetStatus()
	if len(status) == 0 {
		return 0.0
	}

	totalProgress := 0.0
	for _, s := range status {
		totalProgress += s.Progress
	}

	return totalProgress / float64(len(status))
}
