package registry

import (
	"stackyard/pkg/infrastructure"
)

// Dependencies holds all infrastructure dependencies that services might need
type Dependencies struct {
	RedisManager              *infrastructure.RedisManager
	KafkaManager              *infrastructure.KafkaManager
	PostgresManager           *infrastructure.PostgresManager
	PostgresConnectionManager *infrastructure.PostgresConnectionManager
	MongoManager              *infrastructure.MongoManager
	MongoConnectionManager    *infrastructure.MongoConnectionManager
	GrafanaManager            *infrastructure.GrafanaManager
	CronManager               *infrastructure.CronManager
	MinIOManager              *infrastructure.MinIOManager
}

// NewDependencies creates a new dependencies container
func NewDependencies(
	redisManager *infrastructure.RedisManager,
	kafkaManager *infrastructure.KafkaManager,
	postgresManager *infrastructure.PostgresManager,
	postgresConnectionManager *infrastructure.PostgresConnectionManager,
	mongoManager *infrastructure.MongoManager,
	mongoConnectionManager *infrastructure.MongoConnectionManager,
	grafanaManager *infrastructure.GrafanaManager,
	cronManager *infrastructure.CronManager,
	minIOManager *infrastructure.MinIOManager,
) *Dependencies {
	return &Dependencies{
		RedisManager:              redisManager,
		KafkaManager:              kafkaManager,
		PostgresManager:           postgresManager,
		PostgresConnectionManager: postgresConnectionManager,
		MongoManager:              mongoManager,
		MongoConnectionManager:    mongoConnectionManager,
		GrafanaManager:            grafanaManager,
		CronManager:               cronManager,
		MinIOManager:              minIOManager,
	}
}
