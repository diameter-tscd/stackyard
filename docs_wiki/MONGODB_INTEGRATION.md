# MongoDB Integration Guide

This document provides comprehensive information about MongoDB integration in the project, including configuration, usage patterns, and implementation details.

## Overview

MongoDB has been integrated into the project as a full-featured NoSQL database option alongside PostgreSQL. The implementation supports multiple database connections with dynamic switching, similar to the PostgreSQL multi-tenant architecture.

## Features

- **Multiple Database Connections**: Support for multiple MongoDB databases with connection switching
- **Web Monitoring Integration**: Full integration with the web monitoring dashboard
- **CRUD Operations**: Complete Create, Read, Update, Delete operations
- **Aggregation Support**: MongoDB aggregation pipeline operations
- **Connection Health Monitoring**: Real-time connection status and database statistics
- **Manual Query Execution**: Execute raw MongoDB queries through the web interface

## Configuration

### Basic Configuration

Add MongoDB connections to your `config.yaml`:

```yaml
mongo:
  enabled: true
  connections:
    - name: "primary"
      enabled: true
      uri: "mongodb://localhost:27017"
      database: "primary_db"

    - name: "secondary"
      enabled: true
      uri: "mongodb://localhost:27018"
      database: "secondary_db"

    - name: "analytics"
      enabled: false  # Disabled by default
      uri: "mongodb://analytics.example.com:27017"
      database: "analytics_db"
```

### Connection Parameters

| Parameter | Type | Description | Required |
|-----------|------|-------------|----------|
| `name` | string | Connection identifier | Yes |
| `enabled` | boolean | Enable/disable this connection | Yes |
| `uri` | string | MongoDB connection URI | Yes |
| `database` | string | Default database name | Yes |

### URI Format

MongoDB URIs follow the standard format:
```
mongodb://[username:password@]host[:port][/database][?options]
```

Examples:
- `mongodb://localhost:27017` - Local single instance
- `mongodb://user:pass@host:27017/db?replicaSet=rs0` - With authentication and replica set

## Infrastructure Implementation

### MongoDB Manager

The `MongoManager` provides a high-level interface to MongoDB:

```go
type MongoManager struct {
    Client   *mongo.Client
    Database *mongo.Database
}
```

### Connection Manager

The `MongoConnectionManager` handles multiple connections:

```go
type MongoConnectionManager struct {
    connections map[string]*MongoManager
    mu          sync.RWMutex
}
```

### Key Methods

#### Connection Management
```go
// Get a specific connection
conn, exists := mongoManager.GetConnection("primary")

// Get default connection
defaultConn, exists := mongoManager.GetDefaultConnection()

// Get all connections
allConnections := mongoManager.GetAllConnections()

// Get connection status
status := mongoManager.GetStatus()
```

#### Database Operations
```go
// Insert operations
result, err := mongoManager.InsertOne(ctx, "collection", document)
result, err := mongoManager.InsertMany(ctx, "collection", documents)

// Query operations
cursor, err := mongoManager.Find(ctx, "collection", filter)
singleResult := mongoManager.FindOne(ctx, "collection", filter)

// Update operations
result, err := mongoManager.UpdateOne(ctx, "collection", filter, update)
result, err := mongoManager.UpdateMany(ctx, "collection", filter, update)

// Delete operations
result, err := mongoManager.DeleteOne(ctx, "collection", filter)
result, err := mongoManager.DeleteMany(ctx, "collection", filter)

// Aggregation
cursor, err := mongoManager.Aggregate(ctx, "collection", pipeline)

// Utility operations
count, err := mongoManager.CountDocuments(ctx, "collection", filter)
collections, err := mongoManager.ListCollections(ctx)
```

## Web Monitoring Integration

### MongoDB Tab

The MongoDB tab provides a complete interface for database management:

#### Connection Selector
- Dropdown to switch between configured MongoDB connections
- Real-time connection status indicators
- Automatic refresh of database information

#### Database Statistics Cards
- **Database**: Current database name
- **Collections**: Number of collections and list
- **Documents**: Total document count across collections
- **Storage Size**: Database storage usage in MB

#### Manual Query Interface
- **Collection Field**: Specify target collection
- **Query Field**: JSON query/filter (e.g., `{"status": "active"}`)
- **Run Query Button**: Execute the query with loading indicator
- **Results Table**: Dynamic table showing query results

### API Endpoints

#### Get Database Info
```http
GET /api/mongo/info?connection=primary
```

Response:
```json
{
  "success": true,
  "data": {
    "database": "primary_db",
    "collections": ["users", "orders", "products"],
    "stats": {
      "collections": 3,
      "objects": 1250,
      "dataSize": 5242880,
      "storageSize": 8388608,
      "indexes": 5,
      "indexSize": 204800
    }
  }
}
```

#### Execute Query
```http
POST /api/mongo/query?connection=primary
Content-Type: application/json

{
  "collection": "users",
  "query": {"status": "active"}
}
```

Response:
```json
{
  "success": true,
  "data": [
    {
      "_id": "507f1f77bcf86cd799439011",
      "username": "john_doe",
      "email": "john@example.com",
      "status": "active"
    }
  ]
}
```

## Usage Examples

### Basic CRUD Operations

```go
// Inject MongoDB manager
type UserService struct {
    mongoManager *infrastructure.MongoManager
}

func (s *UserService) CreateUser(ctx context.Context, user User) error {
    _, err := s.mongoManager.InsertOne(ctx, "users", user)
    return err
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
    filter := bson.M{"_id": userID}
    var user User
    err := s.mongoManager.FindOne(ctx, "users", filter).Decode(&user)
    return &user, err
}

func (s *UserService) UpdateUser(ctx context.Context, userID string, updates bson.M) error {
    filter := bson.M{"_id": userID}
    _, err := s.mongoManager.UpdateOne(ctx, "users", filter, bson.M{"$set": updates})
    return err
}

func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
    filter := bson.M{"_id": userID}
    _, err := s.mongoManager.DeleteOne(ctx, "users", filter)
    return err
}
```

### Multi-Database Operations

```go
// Using connection manager for multi-tenant operations
type MultiTenantService struct {
    mongoConnectionManager *infrastructure.MongoConnectionManager
}

func (s *MultiTenantService) GetTenantData(ctx context.Context, tenantID string, collection string) ([]bson.M, error) {
    // Get tenant-specific database connection
    conn, exists := s.mongoConnectionManager.GetConnection(tenantID)
    if !exists {
        return nil, fmt.Errorf("tenant database not found: %s", tenantID)
    }

    // Query tenant's database
    cursor, err := conn.Find(ctx, collection, bson.M{})
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var results []bson.M
    if err := cursor.All(ctx, &results); err != nil {
        return nil, err
    }

    return results, nil
}
```

### Aggregation Pipeline

```go
func (s *UserService) GetUserStats(ctx context.Context) ([]bson.M, error) {
    pipeline := []bson.M{
        {
            "$group": bson.M{
                "_id": "$status",
                "count": bson.M{"$sum": 1},
            },
        },
        {
            "$sort": bson.M{"count": -1},
        },
    }

    cursor, err := s.mongoManager.Aggregate(ctx, "users", pipeline)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var results []bson.M
    if err := cursor.All(ctx, &results); err != nil {
        return nil, err
    }

    return results, nil
}
```

## Service Implementation

### Creating a MongoDB Service

Create a new service file `internal/services/modules/service_g.go`:

```go
package modules

import (
    "context"
    "fmt"
    "test-go/pkg/infrastructure"
    "test-go/pkg/logger"
    "test-go/pkg/response"

    "github.com/labstack/echo/v4"
    "go.mongodb.org/mongo-driver/bson"
)

type ServiceG struct {
    enabled                   bool
    mongoConnectionManager    *infrastructure.MongoConnectionManager
    logger                    *logger.Logger
}

func NewServiceG(
    mongoConnectionManager *infrastructure.MongoConnectionManager,
    enabled bool,
    logger *logger.Logger,
) *ServiceG {
    return &ServiceG{
        enabled:                enabled,
        mongoConnectionManager: mongoConnectionManager,
        logger:                 logger,
    }
}

func (s *ServiceG) Name() string        { return "Service G (MongoDB Products)" }
func (s *ServiceG) Enabled() bool       { return s.enabled && s.mongoConnectionManager != nil }
func (s *ServiceG) Endpoints() []string { return []string{"/products/{tenant}"} }

func (s *ServiceG) RegisterRoutes(g *echo.Group) {
    sub := g.Group("/products")

    // Routes with tenant parameter for database selection
    sub.GET("/:tenant", s.listProductsByTenant)
    sub.POST("/:tenant", s.createProduct)
    sub.GET("/:tenant/:id", s.getProductByTenant)
    sub.PUT("/:tenant/:id", s.updateProduct)
    sub.DELETE("/:tenant/:id", s.deleteProduct)
}

func (s *ServiceG) listProductsByTenant(c echo.Context) error {
    tenant := c.Param("tenant")

    // Get the database connection for this tenant
    dbConn, exists := s.mongoConnectionManager.GetConnection(tenant)
    if !exists {
        return response.NotFound(c, fmt.Sprintf("Tenant database '%s' not found or not connected", tenant))
    }

    // Query products from the tenant's database
    cursor, err := dbConn.Find(context.Background(), "products", bson.M{})
    if err != nil {
        return response.InternalServerError(c, fmt.Sprintf("Failed to query tenant '%s' database: %v", tenant, err))
    }
    defer cursor.Close(context.Background())

    var products []bson.M
    if err := cursor.All(context.Background(), &products); err != nil {
        return response.InternalServerError(c, fmt.Sprintf("Failed to decode products: %v", err))
    }

    return response.Success(c, products, fmt.Sprintf("Products retrieved from tenant '%s' database", tenant))
}

func (s *ServiceG) createProduct(c echo.Context) error {
    tenant := c.Param("tenant")

    // Get the database connection for this tenant
    dbConn, exists := s.mongoConnectionManager.GetConnection(tenant)
    if !exists {
        return response.NotFound(c, fmt.Sprintf("Tenant database '%s' not found or not connected", tenant))
    }

    var product bson.M
    if err := c.Bind(&product); err != nil {
        return response.BadRequest(c, "Invalid product data")
    }

    // Insert into tenant's database
    result, err := dbConn.InsertOne(context.Background(), "products", product)
    if err != nil {
        return response.InternalServerError(c, fmt.Sprintf("Failed to create product in tenant '%s' database: %v", tenant, err))
    }

    // Add the generated ID to the response
    product["_id"] = result.InsertedID

    return response.Created(c, product, fmt.Sprintf("Product created in tenant '%s' database", tenant))
}
```

### Registering the Service

Add to `internal/server/server.go`:

```go
// Add MongoDB service
registry.Register(modules.NewServiceG(s.mongoConnectionManager, s.config.Services.IsEnabled("service_g"), s.logger))
```

Add to `config.yaml`:

```yaml
services:
  service_g: true  # Enable MongoDB products service
```

## Manual Query Examples

### Find Documents
```json
{
  "collection": "users",
  "query": {"status": "active"}
}
```

### Aggregation Query
```json
{
  "collection": "orders",
  "query": {
    "$group": {
      "_id": "$status",
      "count": {"$sum": 1}
    }
  }
}
```

### Complex Filter
```json
{
  "collection": "products",
  "query": {
    "price": {"$gte": 100},
    "category": "electronics"
  }
}
```

## Best Practices

### Connection Management
1. **Use Connection Manager**: Always use `MongoConnectionManager` for multi-database scenarios
2. **Connection Validation**: Check connection existence before operations
3. **Error Handling**: Implement proper error handling for connection failures
4. **Resource Cleanup**: Close cursors and handle context cancellation

### Query Optimization
1. **Indexes**: Create appropriate indexes for frequently queried fields
2. **Projection**: Use projection to limit returned fields
3. **Pagination**: Implement pagination for large result sets
4. **Timeouts**: Set appropriate timeouts for long-running operations

### Security
1. **Authentication**: Use MongoDB authentication in production
2. **SSL/TLS**: Enable SSL for production deployments
3. **Access Control**: Implement proper database user permissions
4. **Input Validation**: Validate all user inputs to prevent injection

### Performance
1. **Connection Pooling**: MongoDB driver handles connection pooling automatically
2. **Batch Operations**: Use `InsertMany` and `UpdateMany` for bulk operations
3. **Aggregation Pipeline**: Use aggregation for complex data processing
4. **Indexing Strategy**: Design indexes based on query patterns

## Monitoring and Troubleshooting

### Connection Status
- Check `/api/mongo/info` endpoint for database information
- Monitor connection status in the web dashboard
- View real-time statistics and collection counts

### Query Debugging
- Use manual query interface for testing queries
- Check MongoDB logs for slow queries
- Monitor query execution times

### Common Issues

**Connection Refused**
- Verify MongoDB is running and accessible
- Check connection URI format
- Ensure network connectivity

**Authentication Failed**
- Verify username/password in URI
- Check user permissions in MongoDB
- Ensure database user exists

**Query Timeout**
- Increase timeout values for long-running queries
- Optimize query performance with indexes
- Consider query restructuring

## Migration Guide

### From Single to Multi-Database

1. **Update Configuration**: Convert single MongoDB config to multi-connection format
2. **Update Services**: Modify services to use `MongoConnectionManager`
3. **Test Connections**: Verify all tenant databases are accessible
4. **Migrate Data**: Move existing data to appropriate tenant databases

### From Other Databases

1. **Export Data**: Export data from source database
2. **Transform Schema**: Adapt schema for MongoDB document structure
3. **Import Data**: Use `mongoimport` or custom scripts
4. **Update Application**: Modify application code to use MongoDB operations
5. **Test Functionality**: Verify all features work with new database

## Integration with Other Components

### Redis Caching
```go
// Cache MongoDB query results in Redis
func (s *Service) GetCachedProducts(ctx context.Context, category string) ([]Product, error) {
    cacheKey := fmt.Sprintf("products:category:%s", category)

    // Try Redis cache first
    if cached, err := s.redis.Get(ctx, cacheKey); err == nil {
        var products []Product
        if err := json.Unmarshal([]byte(cached), &products); err == nil {
            return products, nil
        }
    }

    // Query MongoDB
    products, err := s.getProductsFromMongo(ctx, category)
    if err != nil {
        return nil, err
    }

    // Cache results
    if data, err := json.Marshal(products); err == nil {
        s.redis.Set(ctx, cacheKey, string(data), time.Hour)
    }

    return products, nil
}
```

### Kafka Integration
```go
// Publish MongoDB change events to Kafka
func (s *Service) PublishProductUpdate(ctx context.Context, productID string, update bson.M) error {
    // Update MongoDB
    if err := s.updateProductInMongo(ctx, productID, update); err != nil {
        return err
    }

    // Publish event to Kafka
    event := map[string]interface{}{
        "event_type": "product_updated",
        "product_id": productID,
        "update":     update,
        "timestamp":  time.Now(),
    }

    eventData, _ := json.Marshal(event)
    return s.kafka.Publish("product-events", eventData)
}
```

This MongoDB integration provides a robust, scalable, and feature-rich NoSQL database solution that seamlessly integrates with the existing project architecture.
