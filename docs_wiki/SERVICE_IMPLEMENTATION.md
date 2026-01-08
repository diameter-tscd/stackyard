# Service Implementation Guide

This guide documents how to create, implement, and register services in the boilerplate. Services are modular components that encapsulate business logic and expose HTTP endpoints.

---

## Table of Contents

1. [Overview](#overview)
2. [Service Interface](#service-interface)
3. [Creating a Basic Service](#creating-a-basic-service)
4. [Creating a Service with Dependencies](#creating-a-service-with-dependencies)
5. [Registering the Service](#registering-the-service)
6. [Configuration](#configuration)
7. [Complete Example](#complete-example)

---

## Overview

The service architecture follows these principles:

- **Modularity**: Each service is self-contained and can be enabled/disabled via configuration
- **Interface-based**: All services implement the `Service` interface
- **Dynamic Configuration**: Services are registered in a map, so adding new services requires minimal code changes
- **Dependency Injection**: Services can receive infrastructure dependencies (Redis, Postgres, etc.)

### Directory Structure

```
internal/
  services/
    services.go        # Service interface and registry
    modules/
      service_a.go     # Individual service implementations
      service_b.go
      ...
```

---

## Service Interface

All services must implement the `Service` interface defined in `internal/services/services.go`:

```go
type Service interface {
    Name() string                      // Human-readable name for logging/monitoring
    RegisterRoutes(g *echo.Group)      // Register HTTP routes
    Enabled() bool                     // Whether the service is active
    Endpoints() []string               // List of endpoints (for monitoring UI)
}
```

| Method | Purpose |
|--------|---------|
| `Name()` | Returns a display name shown in logs and monitoring dashboard |
| `RegisterRoutes()` | Registers all HTTP endpoints under the provided Echo group |
| `Enabled()` | Returns whether the service should be active (based on config) |
| `Endpoints()` | Returns a list of endpoint paths for the monitoring UI |

---

## Creating a Basic Service

### Step 1: Create the Service File

Create a new file in `internal/services/modules/`. For example, `service_orders.go`:

```go
package modules

import (
    "your-module/pkg/response"

    "github.com/labstack/echo/v4"
)

type OrdersService struct {
    enabled bool
}

func NewOrdersService(enabled bool) *OrdersService {
    return &OrdersService{enabled: enabled}
}

func (s *OrdersService) Name() string        { return "Orders Service" }
func (s *OrdersService) Enabled() bool       { return s.enabled }
func (s *OrdersService) Endpoints() []string { return []string{"/orders", "/orders/:id"} }

func (s *OrdersService) RegisterRoutes(g *echo.Group) {
    sub := g.Group("/orders")
    
    sub.GET("", s.listOrders)
    sub.GET("/:id", s.getOrder)
    sub.POST("", s.createOrder)
}

// Handler implementations
func (s *OrdersService) listOrders(c echo.Context) error {
    // Your business logic here
    return response.Success(c, []string{"order1", "order2"})
}

func (s *OrdersService) getOrder(c echo.Context) error {
    id := c.Param("id")
    return response.Success(c, map[string]string{"id": id, "status": "pending"})
}

func (s *OrdersService) createOrder(c echo.Context) error {
    // Bind request, validate, create order
    return response.Created(c, map[string]string{"id": "new-order-123"})
}
```

### Key Points

1. The struct stores the `enabled` flag passed from configuration
2. `Name()` returns a human-readable name for logs and monitoring
3. `Endpoints()` lists the base paths for monitoring UI display
4. `RegisterRoutes()` sets up all HTTP handlers under a sub-group

---

## Creating a Service with Multiple Database Connections

### Multi-Tenant Services with GORM

For services that need to work with multiple PostgreSQL databases (multi-tenancy), inject the `PostgresConnectionManager`:

```go
package modules

import (
    "fmt"
    "strconv"
    "test-go/pkg/infrastructure"
    "test-go/pkg/response"

    "github.com/labstack/echo/v4"
    "gorm.io/gorm"
)

type MultiTenantOrder struct {
    gorm.Model
    TenantID    string  `json:"tenant_id" gorm:"not null;index"`
    CustomerID  uint    `json:"customer_id" gorm:"not null"`
    ProductName string  `json:"product_name" gorm:"not null"`
    Quantity    int     `json:"quantity" gorm:"not null;check:quantity > 0"`
    TotalPrice  float64 `json:"total_price" gorm:"not null;type:decimal(10,2)"`
    Status      string  `json:"status" gorm:"not null;default:'pending'"`
}

type OrdersService struct {
    enabled                   bool
    postgresConnectionManager *infrastructure.PostgresConnectionManager
}

func NewOrdersService(
    enabled bool,
    postgresConnectionManager *infrastructure.PostgresConnectionManager,
) *OrdersService {
    // Auto-migrate schema for all connected databases
    if enabled && postgresConnectionManager != nil {
        allConnections := postgresConnectionManager.GetAllConnections()
        for tenant, db := range allConnections {
            if db.ORM != nil {
                if err := db.ORM.AutoMigrate(&MultiTenantOrder{}); err != nil {
                    fmt.Printf("Error migrating MultiTenantOrder for tenant '%s': %v\n", tenant, err)
                }
            }
        }
    }

    return &OrdersService{
        enabled:                   enabled,
        postgresConnectionManager: postgresConnectionManager,
    }
}

func (s *OrdersService) Name() string        { return "Multi-Tenant Orders Service" }
func (s *OrdersService) Enabled() bool       { return s.enabled && s.postgresConnectionManager != nil }
func (s *OrdersService) Endpoints() []string { return []string{"/orders/{tenant}", "/orders/{tenant}/{id}"} }

func (s *OrdersService) RegisterRoutes(g *echo.Group) {
    sub := g.Group("/orders")

    // Routes with tenant parameter for database selection
    sub.GET("/:tenant", s.listOrdersByTenant)
    sub.POST("/:tenant", s.createOrder)
    sub.GET("/:tenant/:id", s.getOrderByTenant)
    sub.PUT("/:tenant/:id", s.updateOrder)
    sub.DELETE("/:tenant/:id", s.deleteOrder)
}

func (s *OrdersService) listOrdersByTenant(c echo.Context) error {
    tenant := c.Param("tenant")

    // Get the database connection for this tenant
    dbConn, exists := s.postgresConnectionManager.GetConnection(tenant)
    if !exists {
        return response.NotFound(c, fmt.Sprintf("Tenant database '%s' not found or not connected", tenant))
    }

    // Query using GORM
    var orders []MultiTenantOrder
    result := dbConn.ORM.Where("tenant_id = ?", tenant).Order("created_at DESC").Find(&orders)
    if result.Error != nil {
        return response.InternalServerError(c, fmt.Sprintf("Failed to query tenant '%s' database: %v", tenant, result.Error))
    }

    return response.Success(c, orders, fmt.Sprintf("Orders retrieved from tenant '%s' database", tenant))
}

func (s *OrdersService) createOrder(c echo.Context) error {
    tenant := c.Param("tenant")

    // Get the database connection for this tenant
    dbConn, exists := s.postgresConnectionManager.GetConnection(tenant)
    if !exists {
        return response.NotFound(c, fmt.Sprintf("Tenant database '%s' not found or not connected", tenant))
    }

    var order MultiTenantOrder
    if err := c.Bind(&order); err != nil {
        return response.BadRequest(c, "Invalid order data")
    }

    // Set tenant ID and create using GORM
    order.TenantID = tenant
    order.Status = "pending"

    result := dbConn.ORM.Create(&order)
    if result.Error != nil {
        return response.InternalServerError(c, fmt.Sprintf("Failed to create order in tenant '%s' database: %v", tenant, result.Error))
    }

    return response.Created(c, order, fmt.Sprintf("Order created in tenant '%s' database", tenant))
}
```

### Benefits of Multi-Tenant Architecture

- **Data Isolation**: Each tenant's data is completely separated
- **Scalability**: Different tenants can use different database instances
- **Performance**: Queries are isolated to specific tenant databases
- **Security**: Tenant data cannot accidentally mix
- **Flexibility**: Tenants can have different database configurations

### Configuration for Multi-Tenant Services

```yaml
postgres:
  enabled: true
  connections:
    - name: "tenant_a"
      enabled: true
      host: "localhost"
      port: 5432
      user: "postgres"
      password: "password"
      dbname: "tenant_a_db"
      sslmode: "disable"

    - name: "tenant_b"
      enabled: true
      host: "localhost"
      port: 5433
      user: "postgres"
      password: "password"
      dbname: "tenant_b_db"
      sslmode: "disable"
```

## Creating a Service with Dependencies

For services that require infrastructure (database, cache, etc.), inject dependencies via the constructor:

```go
package modules

import (
    "your-module/pkg/infrastructure"
    "your-module/pkg/response"

    "github.com/labstack/echo/v4"
)

type InventoryService struct {
    db      *infrastructure.PostgresManager
    redis   *infrastructure.RedisManager
    enabled bool
}

func NewInventoryService(
    db *infrastructure.PostgresManager,
    redis *infrastructure.RedisManager,
    enabled bool,
) *InventoryService {
    return &InventoryService{
        db:      db,
        redis:   redis,
        enabled: enabled,
    }
}

func (s *InventoryService) Name() string { return "Inventory Service" }

func (s *InventoryService) Enabled() bool {
    // Can add additional checks for required dependencies
    return s.enabled && s.db != nil
}

func (s *InventoryService) Endpoints() []string {
    return []string{"/inventory"}
}

func (s *InventoryService) RegisterRoutes(g *echo.Group) {
    sub := g.Group("/inventory")
    sub.GET("", s.getInventory)
    sub.PUT("/:sku", s.updateStock)
}

func (s *InventoryService) getInventory(c echo.Context) error {
    // Use s.db or s.redis for data operations
    return response.Success(c, nil)
}

func (s *InventoryService) updateStock(c echo.Context) error {
    return response.Success(c, nil)
}
```

### Conditional Enabling

The `Enabled()` method can include dependency checks:

```go
func (s *InventoryService) Enabled() bool {
    // Only enable if config says enabled AND database is available
    return s.enabled && s.db != nil && s.db.ORM != nil
}
```

---

## Registering the Service

### Step 2: Add to Service List (That's It!)

Adding a new service is now incredibly simple. Just add your service to the list in `internal/services/register.go`:

```go
{
    Name: "orders",
    Constructor: func() interface{ Service } {
        return modules.NewOrdersService(sr.config.Services.IsEnabled("orders"))
    },
},
```

### How the Simplified System Works

The registration system has been completely redesigned for simplicity:

#### 1. **Clean Service List**
All services are defined in a simple array - no complex dependency detection or goroutines:

```go
services := []ServiceDefinition{
    {
        Name: "service_a",
        Constructor: func() interface{ Service } {
            return modules.NewServiceA(sr.config.Services.IsEnabled("service_a"))
        },
    },
    // Add your new service here - that's it!
    {
        Name: "your_service",
        Constructor: func() interface{ Service } {
            return modules.NewYourService(sr.config.Services.IsEnabled("your_service"))
        },
    },
}
```

#### 2. **Synchronous Registration**
Services are registered and booted immediately:

```go
// Register and boot all services
for _, svc := range services {
    service := svc.Constructor()
    registry.Register(service)
    sr.logger.Info("Registered service", "service", svc.Name)
}

registry.Boot(echo)
sr.logger.Info("All services registered and booted successfully")
```

### Server Integration (Clean and Simple)

The server integration is now straightforward:

```go
// internal/server/server.go - Simple service registration
serviceRegistrar := services.NewServiceRegistrar(
    s.config, s.logger,
    s.redisManager, s.kafkaManager, s.postgresManager,
    s.postgresConnectionManager, s.mongoManager,
    s.mongoConnectionManager, s.cronManager,
)

// Register all services (simple and straightforward)
serviceRegistrar.RegisterAllServices(registry, s.echo)
s.logger.Info("All services registered successfully, ready to start monitoring")
```

### Service Key Convention

The string passed to `IsEnabled()` is the key used in `config.yaml`:

| Code | Config Key |
|------|------------|
| `IsEnabled("orders")` | `services.orders` |
| `IsEnabled("inventory")` | `services.inventory` |
| `IsEnabled("service_a")` | `services.service_a` |

---

## Configuration

### Step 3: Add to config.yaml

Add your service key to the `services` section:

```yaml
services:
  service_a: true
  service_b: false
  service_c: true
  service_d: false
  orders: true           # Your new service
  inventory: true        # Another new service
```

### Configuration Behavior

| Value | Behavior |
|-------|----------|
| `true` | Service is enabled |
| `false` | Service is disabled (skipped at startup) |
| Not specified | Defaults to `true` (enabled) |

The default-to-enabled behavior is defined in `config/config.go`:

```go
func (s ServicesConfig) IsEnabled(serviceName string) bool {
    if enabled, exists := s[serviceName]; exists {
        return enabled
    }
    return true // Default to enabled if not specified
}
```

---

## Complete Example

Here is a complete walkthrough for adding a new "Notifications" service:

### 1. Create the Service File

`internal/services/modules/notifications.go`:

```go
package modules

import (
    "your-module/pkg/response"

    "github.com/labstack/echo/v4"
)

type NotificationsService struct {
    enabled bool
}

func NewNotificationsService(enabled bool) *NotificationsService {
    return &NotificationsService{enabled: enabled}
}

func (s *NotificationsService) Name() string        { return "Notifications Service" }
func (s *NotificationsService) Enabled() bool       { return s.enabled }
func (s *NotificationsService) Endpoints() []string { return []string{"/notifications"} }

func (s *NotificationsService) RegisterRoutes(g *echo.Group) {
    sub := g.Group("/notifications")
    
    sub.GET("", func(c echo.Context) error {
        return response.Success(c, []map[string]string{
            {"id": "1", "message": "Welcome!", "read": "false"},
            {"id": "2", "message": "New update available", "read": "true"},
        })
    })
    
    sub.POST("/:id/read", func(c echo.Context) error {
        id := c.Param("id")
        return response.Success(c, nil, "Notification "+id+" marked as read")
    })
}
```

### 2. Register the Service

`internal/server/server.go`:

```go
// In the Start() method, add after existing services:
registry.Register(modules.NewNotificationsService(s.config.Services.IsEnabled("notifications")))
```

### 3. Configure

`config.yaml`:

```yaml
services:
  service_a: true
  service_b: false
  notifications: true    # Enable the new service
```

### 4. Test

Start the application and verify:

- Check the startup logs for "Starting Service... Notifications Service"
- Access `GET /api/v1/notifications`
- Check the monitoring dashboard at `http://localhost:9090` to see the service listed

---

## Summary Checklist

When adding a new service:

1. Create a new file in `internal/services/modules/`
2. Implement the `Service` interface (Name, Enabled, Endpoints, RegisterRoutes)
3. Create a constructor that accepts `enabled bool` (and any dependencies)
4. Register the service in `internal/server/server.go` using `IsEnabled("key")`
5. Add the service key to `config.yaml` under `services:`

No changes to `config/config.go` are required since the services configuration is dynamic.
