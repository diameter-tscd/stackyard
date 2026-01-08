package services

import (
	"test-go/config"
	"test-go/internal/services/modules"
	"test-go/pkg/infrastructure"
	"test-go/pkg/logger"

	"github.com/labstack/echo/v4"
)

// ServiceDefinition holds service registration information
type ServiceDefinition struct {
	Name        string
	Constructor func() interface{ Service }
}

// ServiceRegistrar handles service registration
type ServiceRegistrar struct {
	config          *config.Config
	logger          *logger.Logger
	redisManager    *infrastructure.RedisManager
	kafkaManager    *infrastructure.KafkaManager
	postgresManager *infrastructure.PostgresManager
	postgresConnMgr *infrastructure.PostgresConnectionManager
	mongoManager    *infrastructure.MongoManager
	mongoConnMgr    *infrastructure.MongoConnectionManager
	grafanaManager  *infrastructure.GrafanaManager
	cronManager     *infrastructure.CronManager
}

// NewServiceRegistrar creates a new service registrar
func NewServiceRegistrar(
	cfg *config.Config,
	logger *logger.Logger,
	redisMgr *infrastructure.RedisManager,
	kafkaMgr *infrastructure.KafkaManager,
	postgresMgr *infrastructure.PostgresManager,
	postgresConnMgr *infrastructure.PostgresConnectionManager,
	mongoMgr *infrastructure.MongoManager,
	mongoConnMgr *infrastructure.MongoConnectionManager,
	grafanaMgr *infrastructure.GrafanaManager,
	cronMgr *infrastructure.CronManager,
) *ServiceRegistrar {
	return &ServiceRegistrar{
		config:          cfg,
		logger:          logger,
		redisManager:    redisMgr,
		kafkaManager:    kafkaMgr,
		postgresManager: postgresMgr,
		postgresConnMgr: postgresConnMgr,
		mongoManager:    mongoMgr,
		mongoConnMgr:    mongoConnMgr,
		grafanaManager:  grafanaMgr,
		cronManager:     cronMgr,
	}
}

/*
HOW TO ADD A NEW SERVICE:

1. Create your service file in internal/services/modules/ (e.g., service_orders.go)
2. Implement the Service interface (Name, Enabled, Endpoints, RegisterRoutes)
3. Add your service to the list below - that's it!

EXAMPLE:

// In internal/services/modules/service_orders.go
type OrdersService struct {
	enabled bool
}

func NewOrdersService(enabled bool) *OrdersService {
	return &OrdersService{enabled: enabled}
}

func (s *OrdersService) Name() string        { return "Orders Service" }
func (s *OrdersService) Enabled() bool       { return s.enabled }
func (s *OrdersService) Endpoints() []string { return []string{"/orders"} }

func (s *OrdersService) RegisterRoutes(g *echo.Group) {
	sub := g.Group("/orders")
	sub.GET("", s.listOrders)
	sub.POST("", s.createOrder)
}

// Add to config.yaml under services:
// services:
//   orders: true

// Then add to the list below:
// {
// 	Name: "orders",
// 	Constructor: func() interface{ Service } {
// 		return modules.NewOrdersService(sr.config.Services.IsEnabled("orders"))
// 	},
// },
*/

// RegisterAllServices registers all services
// Just add your new service below - that's it!
func (sr *ServiceRegistrar) RegisterAllServices(registry *Registry, echo *echo.Echo) {
	services := []ServiceDefinition{
		// ===============================
		// ADD YOUR NEW SERVICE HERE
		// ===============================
		{
			Name: "service_a",
			Constructor: func() interface{ Service } {
				return modules.NewServiceA(sr.config.Services.IsEnabled("service_a"))
			},
		},
		{
			Name: "service_b",
			Constructor: func() interface{ Service } {
				return modules.NewServiceB(sr.config.Services.IsEnabled("service_b"))
			},
		},
		{
			Name: "service_c",
			Constructor: func() interface{ Service } {
				return modules.NewServiceC(sr.config.Services.IsEnabled("service_c"))
			},
		},
		{
			Name: "service_d",
			Constructor: func() interface{ Service } {
				return modules.NewServiceD(sr.postgresManager, sr.config.Services.IsEnabled("service_d"), sr.logger)
			},
		},
		{
			Name: "service_e",
			Constructor: func() interface{ Service } {
				encryptionConfig := map[string]interface{}{
					"algorithm": sr.config.Encryption.Algorithm,
					"key":       sr.config.Encryption.Key,
				}
				return modules.NewServiceE(sr.config.Encryption.Enabled, encryptionConfig)
			},
		},
		{
			Name: "service_f",
			Constructor: func() interface{ Service } {
				return modules.NewServiceF(sr.postgresConnMgr, sr.config.Services.IsEnabled("service_f"), sr.logger)
			},
		},
		{
			Name: "service_g",
			Constructor: func() interface{ Service } {
				return modules.NewServiceG(sr.mongoConnMgr, sr.config.Services.IsEnabled("service_g"), sr.logger)
			},
		},
		{
			Name: "service_h",
			Constructor: func() interface{ Service } {
				return modules.NewServiceH(sr.config.Services.IsEnabled("service_h"), sr.logger)
			},
		},
		{
			Name: "service_i",
			Constructor: func() interface{ Service } {
				return modules.NewServiceI(sr.grafanaManager, sr.config.Services.IsEnabled("service_i"), sr.logger)
			},
		},

		// ===============================
		// ADD YOUR NEW SERVICE ABOVE THIS LINE
		// ===============================
	}

	// Register and boot all services
	for _, svc := range services {
		service := svc.Constructor()
		registry.Register(service)
		sr.logger.Info("Registered service", "service", svc.Name)
	}

	registry.Boot(echo)
	sr.logger.Info("All services registered and booted successfully")
}
