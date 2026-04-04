package registry

import (
	"stackyrd/config"
	"stackyrd/pkg/infrastructure"
	"stackyrd/pkg/logger"
)

// ServiceHelper helps services with dependency validation
type ServiceHelper struct {
	config *config.Config
	logger *logger.Logger
	deps   *Dependencies
}

// NewServiceHelper creates a new service helper
func NewServiceHelper(config *config.Config, logger *logger.Logger, deps *Dependencies) *ServiceHelper {
	return &ServiceHelper{
		config: config,
		logger: logger,
		deps:   deps,
	}
}

// RequireDependency validates dependency is available
func (h *ServiceHelper) RequireDependency(name string, available bool) bool {
	if !available {
		h.logger.Warn(name + " not available, skipping service")
		return false
	}
	return true
}

// IsServiceEnabled checks if service is enabled in config
func (h *ServiceHelper) IsServiceEnabled(serviceName string) bool {
	return h.config.Services.IsEnabled(serviceName)
}

// GetRedis returns Redis manager or error if not available
func (h *ServiceHelper) GetRedis() (*infrastructure.RedisManager, bool) {
	return GetTyped[*infrastructure.RedisManager](h.deps, "redis")
}

// GetKafka returns Kafka manager or error if not available
func (h *ServiceHelper) GetKafka() (*infrastructure.KafkaManager, bool) {
	return GetTyped[*infrastructure.KafkaManager](h.deps, "kafka")
}

// GetPostgres returns PostgreSQL manager (single connection) or error
func (h *ServiceHelper) GetPostgres() (*infrastructure.PostgresManager, bool) {
	return GetTyped[*infrastructure.PostgresManager](h.deps, "postgres")
}

// GetPostgresConnection returns PostgreSQL connection manager (multi-tenant) or error
func (h *ServiceHelper) GetPostgresConnection() (*infrastructure.PostgresConnectionManager, bool) {
	return GetTyped[*infrastructure.PostgresConnectionManager](h.deps, "postgres")
}

// GetMongo returns MongoDB manager (single connection) or error
func (h *ServiceHelper) GetMongo() (*infrastructure.MongoManager, bool) {
	return GetTyped[*infrastructure.MongoManager](h.deps, "mongo")
}

// GetMongoConnection returns MongoDB connection manager (multi-tenant) or error
func (h *ServiceHelper) GetMongoConnection() (*infrastructure.MongoConnectionManager, bool) {
	return GetTyped[*infrastructure.MongoConnectionManager](h.deps, "mongo")
}

// GetGrafana returns Grafana manager or error if not available
func (h *ServiceHelper) GetGrafana() (*infrastructure.GrafanaManager, bool) {
	return GetTyped[*infrastructure.GrafanaManager](h.deps, "grafana")
}

// GetCron returns Cron manager or error if not available
func (h *ServiceHelper) GetCron() (*infrastructure.CronManager, bool) {
	return GetTyped[*infrastructure.CronManager](h.deps, "cron")
}

// GetMinIO returns MinIO manager or error if not available
func (h *ServiceHelper) GetMinIO() (*infrastructure.MinIOManager, bool) {
	return GetTyped[*infrastructure.MinIOManager](h.deps, "minio")
}
