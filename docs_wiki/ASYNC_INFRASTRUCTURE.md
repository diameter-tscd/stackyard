# Async Infrastructure Implementation Guide

## Overview

This document describes the async infrastructure implementation that ensures all database operations, caching, message queuing, and file operations run asynchronously to avoid blocking the main application thread. The implementation uses Go's goroutines, channels, and worker pools to provide non-blocking operations while maintaining thread safety.

The system also includes **async infrastructure initialization** that allows the HTTP server to start immediately without waiting for database connections and other infrastructure components to initialize. This provides instant responsiveness while infrastructure components initialize in the background.

## Key Components

### 1. AsyncResult Types

The async infrastructure uses generic `AsyncResult[T]` types to handle asynchronous operations:

```go
type AsyncResult[T any] struct {
    Value T
    Error error
    Done  chan struct{}
}
```

**Key Methods:**
- `Wait()` - Blocks until operation completes
- `WaitWithTimeout(timeout)` - Waits with timeout
- `IsDone()` - Non-blocking check if operation is complete

### 2. Worker Pools

Each infrastructure component includes a worker pool for managing concurrent operations:

```go
type WorkerPool struct {
    workers  int
    jobQueue chan func()
    stopChan chan struct{}
    stopped  chan struct{}
}
```

**Benefits:**
- Controlled concurrency
- Resource management
- Panic recovery
- Graceful shutdown

### 3. Batch Operations

Support for batching multiple operations:

```go
type BatchAsyncResult[T any] struct {
    Results []AsyncResult[T]
    Done    chan struct{}
}
```

## Infrastructure Components

### Redis Manager

**Async Operations:**
```go
// Set value asynchronously
result := redisManager.SetAsync(ctx, "key", "value", time.Hour)

// Get value asynchronously
result := redisManager.GetAsync(ctx, "key")

// Wait for result
value, err := result.Wait()
```

**Batch Operations:**
```go
// Set multiple keys
kvPairs := map[string]interface{}{"key1": "value1", "key2": "value2"}
result := redisManager.SetBatchAsync(ctx, kvPairs, time.Hour)

// Get multiple keys
keys := []string{"key1", "key2"}
result := redisManager.GetBatchAsync(ctx, keys)
```

**Worker Pool Integration:**
```go
// Submit background job
redisManager.SubmitAsyncJob(func() {
    // Long-running Redis operation
})
```

### Kafka Manager

**Async Operations:**
```go
// Publish message asynchronously
result := kafkaManager.PublishAsync(ctx, "topic", []byte("message"))

// Publish with key
result := kafkaManager.PublishWithKeyAsync(ctx, "topic", []byte("key"), []byte("message"))
```

**Batch Operations:**
```go
// Publish multiple messages
messages := [][]byte{[]byte("msg1"), []byte("msg2")}
result := kafkaManager.PublishBatchAsync(ctx, "topic", messages)
```

**Consumer Operations:**
```go
// Start consumer asynchronously (doesn't block)
kafkaManager.ConsumeAsync(ctx, "topic", func(key, value []byte) error {
    // Handle message
    return nil
})
```

### MinIO Manager

**Async Operations:**
```go
// Upload file asynchronously
file, _ := os.Open("file.txt")
defer file.Close()
result := minioManager.UploadFileAsync(ctx, "object.txt", file, size, "text/plain")
```

**Batch Operations:**
```go
// Upload multiple files
uploads := []struct{
    ObjectName, Reader, Size, ContentType
}{/* file data */}
result := minioManager.UploadBatchAsync(ctx, uploads)
```

### PostgreSQL Manager

**Async Operations:**
```go
// Execute query asynchronously
result := postgresManager.QueryAsync(ctx, "SELECT * FROM users", args...)

// Execute DML operations
insertResult := postgresManager.InsertAsync(ctx, "INSERT INTO users...", args...)
updateResult := postgresManager.UpdateAsync(ctx, "UPDATE users...", args...)
deleteResult := postgresManager.DeleteAsync(ctx, "DELETE FROM users...", args...)
```

**GORM Async Operations:**
```go
// Async GORM operations
createResult := postgresManager.GORMCreateAsync(ctx, &user)
findResult := postgresManager.GORMFindAsync(ctx, &users)
updateResult := postgresManager.GORMUpdateAsync(ctx, &user, updates, "id = ?", id)
deleteResult := postgresManager.GORMDeleteAsync(ctx, &user, "id = ?", id)
```

### MongoDB Manager

**Async Operations:**
```go
// CRUD operations
insertResult := mongoManager.InsertOneAsync(ctx, "collection", document)
findResult := mongoManager.FindAsync(ctx, "collection", filter)
updateResult := mongoManager.UpdateOneAsync(ctx, "collection", filter, update)
deleteResult := mongoManager.DeleteOneAsync(ctx, "collection", filter)
```

**Batch Operations:**
```go
// Batch insert
inserts := []struct{Collection string; Document interface{}}{/* data */}
result := mongoManager.InsertBatchAsync(ctx, inserts)
```

### Cron Manager

**Async Job Execution:**
```go
// Add job that executes asynchronously in worker pool
cronManager.AddAsyncJob("name", "schedule", func() {
    // Job logic (runs in worker pool, doesn't block main thread)
}))

// Run job immediately (asynchronously)
cronManager.RunJobNow(jobID)

// Get job status and scheduling info
jobs := cronManager.GetJobs() // Returns scheduled jobs with next run times
```

**Configuration:**
```yaml
cron:
  enabled: true
  jobs:
    log_cleanup: "0 0 * * *"        # Daily at midnight
    health_check: "*/10 * * * * *"  # Every 10 seconds
```

**Features:**
- **Async Execution**: Jobs run in worker pools to avoid blocking
- **Schedule Management**: Add, remove, update job schedules dynamically
- **Status Monitoring**: View active jobs and execution history
- **Graceful Shutdown**: Clean termination of running jobs

## Async Infrastructure Initialization

The application implements **async infrastructure initialization** that allows the HTTP server to start immediately without waiting for database connections and other infrastructure components to initialize. This provides instant responsiveness while infrastructure components initialize in the background.

### Infrastructure Initialization Manager

The `InfraInitManager` manages asynchronous initialization of all infrastructure components:

```go
type InfraInitManager struct {
    status   map[string]*InfraInitStatus
    mu       sync.RWMutex
    logger   *logger.Logger
    doneChan chan struct{}
}
```

### Initialization Process

1. **Immediate Server Start**: HTTP server starts immediately without waiting
2. **Background Initialization**: Infrastructure components initialize concurrently in goroutines
3. **Progress Tracking**: Real-time status monitoring of initialization progress
4. **Health Endpoints**: API endpoints provide initialization status

### Usage Example

```go
// In server startup
infraInitManager := infrastructure.NewInfraInitManager(logger)

// Start async initialization (doesn't block)
redisMgr, kafkaMgr, _, postgresMgr, mongoMgr, cronMgr :=
    infraInitManager.StartAsyncInitialization(config, logger)

// HTTP server starts immediately here
// Infrastructure initializes in background

// Check initialization status
status := infraInitManager.GetStatus()
progress := infraInitManager.GetInitializationProgress()
```

### Health Check Endpoints

#### GET /health
Enhanced health check with infrastructure status:
```json
{
  "status": "ok",
  "server_ready": true,
  "infrastructure": {
    "redis": {
      "name": "redis",
      "initialized": true,
      "start_time": "2025-12-19T13:13:00Z",
      "duration": "2.5s",
      "progress": 1.0
    },
    "postgres": {
      "name": "postgres",
      "initialized": true,
      "start_time": "2025-12-19T13:13:00Z",
      "duration": "3.2s",
      "progress": 1.0
    }
  },
  "initialization_progress": 0.85
}
```

#### GET /health/infrastructure
Detailed infrastructure initialization status:
```json
{
  "redis": {
    "name": "redis",
    "initialized": true,
    "error": "",
    "start_time": "2025-12-19T13:13:00Z",
    "duration": "2.5s",
    "progress": 1.0
  },
  "postgres": {
    "name": "postgres",
    "initialized": false,
    "error": "connection timeout",
    "start_time": "2025-12-19T13:13:00Z",
    "duration": "30s",
    "progress": 0.0
  }
}
```

### Benefits of Async Initialization

- **Instant Server Responsiveness**: HTTP server available immediately
- **Graceful Degradation**: Services work with partially initialized infrastructure
- **Better User Experience**: No waiting for database connections on startup
- **Fault Tolerance**: Failed components don't prevent server startup
- **Monitoring Integration**: Real-time visibility into initialization progress

### Initialization Order

Components initialize concurrently but with logical dependencies:

1. **Redis** - Fast cache initialization
2. **PostgreSQL/MongoDB** - Database connections (may take longer)
3. **Kafka** - Message queue connections
4. **MinIO** - Object storage
5. **Cron** - Scheduled jobs

### Clean Service Registration with Automatic Dependency Detection

The application implements **clean service registration** with automatic dependency detection using a dedicated `ServiceRegistrar`. Services are registered **only once** in a central location, and the system automatically determines infrastructure dependencies through reflection analysis.

#### Architecture Overview

**Clean Separation of Concerns:**
- **`internal/services/register.go`**: Service registration logic and dependency analysis
- **`internal/server/server.go`**: Clean server startup without registration complexity

**Key Components:**
- **`ServiceRegistrar`**: Handles all service registration and dependency management
- **`ServiceDefinition`**: Simple struct holding service name and constructor
- **Reflection Analysis**: Automatic dependency detection based on struct field types

#### Service Definition (One Clean Location)

```go
// internal/services/register.go
allServices := []ServiceDefinition{
    {
        Name: "service_a",
        Constructor: func() interface{ Service } {
            return modules.NewServiceA(sr.config.Services.IsEnabled("service_a"))
        },
    },
    {
        Name: "service_d",
        Constructor: func() interface{ Service } {
            return modules.NewServiceD(sr.postgresManager, sr.config.Services.IsEnabled("service_d"), sr.logger)
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
}
```

#### Server Startup (Clean and Simple)

```go
// internal/server/server.go - Clean server startup
serviceRegistrar := services.NewServiceRegistrar(
    s.config, s.logger, s.infraInitManager,
    s.redisManager, s.kafkaManager, s.postgresManager,
    s.postgresConnectionManager, s.mongoManager,
    s.mongoConnectionManager, s.cronManager,
)

// Register ALL services with automatic dependency detection
dependentServiceCount := serviceRegistrar.RegisterAllServices(registry, s.echo)

// Wait for dependent services, then start monitoring
if dependentServiceCount > 0 {
    time.Sleep(time.Duration(dependentServiceCount) * 500 * time.Millisecond)
}
// Start monitoring with complete service list
```

#### Reflection-Based Dependency Detection

The `analyzeConstructorDependencies()` method automatically determines dependencies:

```go
func (sr *ServiceRegistrar) analyzeConstructorDependencies(constructor func() interface{ Service }) []string {
    service := constructor() // Create service instance

    // Use reflection to examine struct fields for infrastructure types
    serviceValue := reflect.ValueOf(service)
    serviceType := serviceValue.Type()

    var dependencies []string
    for i := 0; i < serviceType.NumField(); i++ {
        fieldType := serviceType.Field(i).Type.String()

        switch fieldType {
        case "infrastructure.PostgresConnectionManager":
            if sr.config.Postgres.Enabled || sr.config.PostgresMultiConfig.Enabled {
                dependencies = append(dependencies, "postgres")
            }
        case "infrastructure.MongoConnectionManager":
            if sr.config.Mongo.Enabled || sr.config.MongoMultiConfig.Enabled {
                dependencies = append(dependencies, "mongodb")
            }
        // ... other infrastructure types automatically detected
        }
    }
    return dependencies
}
```

#### Synchronous Registration Process with Complete Synchronization

**Phase 1: Infrastructure-Independent Services**
```go
// Start immediately - no dependencies
for _, svc := range independentServices {
    registry.Register(svc.Constructor())
}
registry.Boot(echo)
```

**Phase 2: Infrastructure-Dependent Services (Synchronous)**
```go
// Use channel to track completion of ALL dependent services
dependentDoneChan := make(chan struct{}, len(dependentServices))

for _, svc := range dependentServices {
    go func(serviceDef DependentServiceDefinition) {
        defer func() { dependentDoneChan <- struct{}{} }()

        // Wait for each required infrastructure component
        for _, dep := range serviceDef.Dependencies {
            for !infraInitManager.IsInitialized(dep) {
                time.Sleep(100 * time.Millisecond)
            }
        }
        // Register service when dependencies are ready
        registry.Register(serviceDef.Constructor())
        registry.BootService(echo, serviceDef.Constructor())
    }(svc)
}

// Wait for ALL dependent services to complete registration
for i := 0; i < len(dependentServices); i++ {
    <-dependentDoneChan  // Blocks until each service completes
}
```

**Phase 3: Monitoring Startup (After All Services)**
```go
// Now ALL services are registered - build complete service list
var servicesList []monitoring.ServiceInfo
for _, srv := range registry.GetServices() {
    servicesList = append(servicesList, monitoring.ServiceInfo{
        Name: srv.Name(),
        Active: srv.Enabled(),
        Endpoints: srv.Endpoints(),
    })
}
go monitoring.Start(config, servicesList) // Complete service list
```

#### Benefits of Clean Architecture

- **Separation of Concerns**: Registration logic separated from server logic
- **Single Definition**: Services defined only once, no duplication
- **Automatic Detection**: Dependencies determined by examining actual code
- **Maintainable**: Easy to add new services - just add to the list
- **Clean Server Code**: Server startup logic remains simple and focused
- **Type-Safe**: Uses Go's type system for reliable dependency detection

#### Adding New Services

To add a new service, simply add it to the `allServices` slice:

```go
{
    Name: "service_new",
    Constructor: func() interface{ Service } {
        return modules.NewService(s.someManager, s.config.Services.IsEnabled("service_new"))
    },
},
// System automatically detects dependencies based on constructor parameters
```

This approach provides a **clean, maintainable, and automatic** service registration system that scales effortlessly as new services and infrastructure components are added.

### Error Handling

Failed initializations are logged but don't prevent server startup:

```go
// Initialization continues even if some components fail
if err != nil {
    status.Error = err.Error()
    logger.Error("Failed to initialize infrastructure component", err, "component", name)
}
```

### Configuration

Infrastructure components initialize based on their configuration settings:

```yaml
redis:
  enabled: true  # Will initialize if enabled

postgres:
  enabled: false # Will skip initialization

mongo:
  enabled: true
  multi_config:
    enabled: true # Will initialize multi-connection setup
```

## Graceful Shutdown

The application implements **graceful shutdown** that properly disconnects all infrastructure components when receiving SIGTERM or SIGINT signals. This ensures clean resource cleanup and prevents data corruption.

### Shutdown Process

1. **Signal Handling**: Application catches SIGTERM/SIGINT signals
2. **Infrastructure Shutdown**: Components shut down in reverse order of initialization
3. **Resource Cleanup**: Connections closed, worker pools stopped
4. **Logging**: Detailed shutdown progress logged
5. **Error Reporting**: Shutdown errors reported but don't prevent completion

### Shutdown Order

Components are shut down in reverse order to ensure dependencies are handled correctly:

1. **Cron Manager** - Stop scheduled jobs first
2. **MongoDB Connections** - Close document database connections
3. **PostgreSQL Connections** - Close relational database connections
4. **Kafka Manager** - Stop message producers/consumers
5. **Redis Manager** - Close cache connections last

### Usage Example

```go
// Signal handling in main.go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

<-sigChan

// Graceful shutdown
err := server.Shutdown(context.Background(), logger)
if err != nil {
    logger.Error("Shutdown completed with errors", err)
} else {
    logger.Info("Shutdown completed successfully")
}
```

### Shutdown Method

The `Server.Shutdown()` method performs orderly shutdown:

```go
func (s *Server) Shutdown(ctx context.Context, logger *logger.Logger) error {
    // 1. Cron Manager
    if s.cronManager != nil {
        logger.Info("Shutting down Cron Manager...")
        if err := s.cronManager.Close(); err != nil {
            return fmt.Errorf("cron shutdown error: %w", err)
        }
    }

    // 2. MongoDB connections
    if s.mongoConnectionManager != nil {
        logger.Info("Shutting down MongoDB connections...")
        if err := s.mongoConnectionManager.CloseAll(); err != nil {
            return fmt.Errorf("mongodb shutdown error: %w", err)
        }
    }

    // 3. PostgreSQL connections
    if s.postgresConnectionManager != nil {
        logger.Info("Shutting down PostgreSQL connections...")
        if err := s.postgresConnectionManager.CloseAll(); err != nil {
            return fmt.Errorf("postgres shutdown error: %w", err)
        }
    }

    // 4. Kafka Manager
    if s.kafkaManager != nil {
        logger.Info("Shutting down Kafka Manager...")
        if err := s.kafkaManager.Close(); err != nil {
            return fmt.Errorf("kafka shutdown error: %w", err)
        }
    }

    // 5. Redis Manager
    if s.redisManager != nil {
        logger.Info("Shutting down Redis Manager...")
        if err := s.redisManager.Close(); err != nil {
            return fmt.Errorf("redis shutdown error: %w", err)
        }
    }

    return nil
}
```

### Benefits of Graceful Shutdown

- **Data Integrity**: Prevents partial writes and corruption
- **Resource Cleanup**: Ensures all connections are properly closed
- **Clean Termination**: No hanging processes or zombie goroutines
- **Monitoring**: Shutdown progress is logged and monitored
- **Kubernetes Compatibility**: Works with container orchestration systems

### Error Handling During Shutdown

Shutdown errors are logged but don't prevent the shutdown process:

```go
if err := component.Close(); err != nil {
    shutdownErrors = append(shutdownErrors, fmt.Errorf("component shutdown error: %w", err))
    logger.Error("Error shutting down component", err)
} else {
    logger.Info("Component shut down successfully")
}
```

### Timeout Handling

Shutdown operations can be given timeouts to prevent hanging:

```go
shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := server.Shutdown(shutdownCtx, logger)
```

### Integration with Orchestration

Works seamlessly with container orchestration systems:

```bash
# Docker graceful shutdown
docker stop --timeout 30 container_name

# Kubernetes graceful termination
terminationGracePeriodSeconds: 30
```

### Testing Shutdown

Test graceful shutdown behavior:

```bash
# Send SIGTERM signal
kill -TERM <pid>

# Or use Ctrl+C in interactive mode
# Application will log shutdown progress and exit cleanly
```

## Usage Patterns

### Synchronous Usage (Blocking)

```go
// Execute async operation and wait for result
result := redisManager.GetAsync(ctx, "key")
value, err := result.Wait()

if err != nil {
    // Handle error
}
// Use value
```

### Asynchronous Usage (Non-blocking)

```go
// Start operation without waiting
result := redisManager.GetAsync(ctx, "key")

// Continue with other work
doOtherWork()

// Check if done later
if result.IsDone() {
    value, err := result.Wait()
    // Handle result
}
```

### Timeout Handling

```go
// Wait with timeout
result := redisManager.GetAsync(ctx, "key")
value, err := result.WaitWithTimeout(5 * time.Second)

if err == context.DeadlineExceeded {
    // Handle timeout
}
```

### Batch Processing

```go
// Process multiple operations in parallel
keys := []string{"key1", "key2", "key3"}
result := redisManager.GetBatchAsync(ctx, keys)

// Wait for all operations to complete
values, errors := result.WaitAll()

for i, value := range values {
    if errors[i] != nil {
        // Handle error for this operation
    }
    // Process value
}
```

### Worker Pool Submission

```go
// Submit long-running tasks to worker pool
redisManager.SubmitAsyncJob(func() {
    // Long-running operation
    heavyComputation()
    redisManager.Set(ctx, "result", computedValue, time.Hour)
})
```

## Service Implementation Examples

### Database Service with Async Operations

```go
type UserService struct {
    db *infrastructure.PostgresManager
}

func (s *UserService) CreateUser(c echo.Context) error {
    var user User
    if err := c.Bind(&user); err != nil {
        return response.BadRequest(c, "Invalid user data")
    }

    // Async database operation
    result := s.db.GORMCreateAsync(context.Background(), &user)

    // Wait for completion
    _, err := result.Wait()
    if err != nil {
        return response.InternalServerError(c, err.Error())
    }

    return response.Created(c, user)
}

func (s *UserService) GetUsers(c echo.Context) error {
    var users []User

    // Async query
    result := s.db.GORMFindAsync(context.Background(), &users)

    // Wait for completion
    _, err := result.Wait()
    if err != nil {
        return response.InternalServerError(c, err.Error())
    }

    return response.Success(c, users)
}
```

### Cache Service with Async Operations

```go
type CacheService struct {
    redis *infrastructure.RedisManager
}

func (s *CacheService) SetMultiple(ctx context.Context, data map[string]interface{}) error {
    // Async batch set
    result := s.redis.SetBatchAsync(ctx, data, time.Hour)

    // Wait for all operations
    _, errors := result.WaitAll()

    // Check for any errors
    for _, err := range errors {
        if err != nil {
            return err
        }
    }

    return nil
}
```

### Message Queue Service

```go
type QueueService struct {
    kafka *infrastructure.KafkaManager
}

func (s *QueueService) PublishEvents(ctx context.Context, events []Event) error {
    // Convert events to messages
    messages := make([][]byte, len(events))
    for i, event := range events {
        data, _ := json.Marshal(event)
        messages[i] = data
    }

    // Async batch publish
    result := s.kafka.PublishBatchAsync(ctx, "events", messages)

    // Wait for completion
    _, errors := result.WaitAll()

    // Check for errors
    for _, err := range errors {
        if err != nil {
            return err
        }
    }

    return nil
}
```

## Performance Benefits

### Non-blocking Operations

- **Main Thread Availability**: HTTP handlers return immediately while operations run in background
- **Concurrent Processing**: Multiple operations can run simultaneously
- **Resource Efficiency**: Better utilization of system resources

### Example Performance Comparison

**Synchronous Approach:**
```
Request → DB Query (2s) → Response
Total: 2 seconds per request
Throughput: 0.5 requests/second
```

**Asynchronous Approach:**
```
Request → Start Async DB Query → Return Response (immediate)
DB Query completes (2s) → Result stored/cached
Total: ~0ms per request (non-blocking)
Throughput: Limited by DB capacity, not request handling
```

### Resource Management

- **Worker Pools**: Control maximum concurrent operations
- **Connection Pooling**: Reuse database connections efficiently
- **Timeout Handling**: Prevent hanging operations
- **Graceful Shutdown**: Clean up resources properly

## Error Handling

### Operation-specific Errors

```go
result := redisManager.GetAsync(ctx, "key")
value, err := result.Wait()

if err != nil {
    switch err {
    case redis.Nil:
        // Key not found
    case context.DeadlineExceeded:
        // Operation timed out
    default:
        // Other error
    }
}
```

### Batch Error Handling

```go
result := redisManager.GetBatchAsync(ctx, keys)
values, errors := result.WaitAll()

for i, err := range errors {
    if err != nil {
        log.Printf("Error getting key %s: %v", keys[i], err)
        // Handle specific error
    } else {
        // Process values[i]
    }
}
```

### Panic Recovery

Async operations include panic recovery:

```go
// Automatic panic handling in ExecuteAsync
go func() {
    defer func() {
        if r := recover(); r != nil {
            result.Complete(zeroValue, fmt.Errorf("operation panicked: %v", r))
        }
    }()
    // Operation logic
}()
```

## Monitoring and Observability

### Status Monitoring

Each manager provides status information including async operation stats:

```go
// Redis status includes pool information
redisStatus := redisManager.GetStatus()
// {"connected": true, "pool_workers": 10, "active_jobs": 3}

// Cron status includes worker pool info
cronStatus := cronManager.GetStatus()
// {"active": true, "jobs": [...], "pool_workers": 5}
```

### Performance Metrics

- **Operation Latency**: Time taken for async operations
- **Queue Depth**: Number of pending operations
- **Error Rates**: Failure rates for different operation types
- **Resource Usage**: Memory and CPU usage by worker pools

## Configuration

### Worker Pool Configuration

```yaml
# Infrastructure worker pool sizes
infrastructure:
  redis:
    workers: 10
  kafka:
    workers: 5
  minio:
    workers: 8
  postgres:
    workers: 15
  mongodb:
    workers: 12
  cron:
    workers: 5
```

### Timeout Configuration

```yaml
# Operation timeouts
infrastructure:
  timeouts:
    redis: 30s
    kafka: 60s
    minio: 300s
    postgres: 30s
    mongodb: 30s
```

## Best Practices

### 1. Context Usage

Always use context for cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result := dbManager.QueryAsync(ctx, "SELECT * FROM users")
```

### 2. Error Handling

Handle errors appropriately:

```go
result := redisManager.GetAsync(ctx, "key")
value, err := result.Wait()

if err == context.Canceled {
    // Request was canceled
    return
}
if err == context.DeadlineExceeded {
    // Operation timed out
    return
}
// Handle other errors
```

### 3. Resource Cleanup

Use defer for cleanup:

```go
result := fileManager.UploadAsync(ctx, "file.txt", reader, size, contentType)
defer func() {
    if reader != nil {
        reader.Close()
    }
}()

// Wait for operation
_, err := result.Wait()
```

### 4. Batch Operations

Use batch operations for multiple items:

```go
// Instead of multiple individual operations
for _, key := range keys {
    redisManager.SetAsync(ctx, key, value, ttl)
}

// Use batch operation
kvPairs := make(map[string]interface{})
for _, key := range keys {
    kvPairs[key] = value
}
redisManager.SetBatchAsync(ctx, kvPairs, ttl)
```

### 5. Worker Pool Sizing

Size worker pools based on load:

```go
// High-throughput service
redisPool := NewWorkerPool(50)

// Low-throughput service
cronPool := NewWorkerPool(5)
```

## Migration Guide

### Converting Synchronous Code

**Before (Synchronous):**
```go
func (s *Service) GetUser(id string) (*User, error) {
    var user User
    err := s.db.Where("id = ?", id).First(&user).Error
    return &user, err
}
```

**After (Asynchronous):**
```go
func (s *Service) GetUser(id string) (*User, error) {
    var user User
    result := s.db.GORMFirstAsync(ctx, &user, id)
    _, err := result.Wait()
    return &user, err
}
```

### Gradual Migration

1. **Identify blocking operations**
2. **Add async versions alongside sync versions**
3. **Update callers gradually**
4. **Remove sync versions after full migration**

## Testing

### Unit Testing Async Operations

```go
func TestAsyncRedisGet(t *testing.T) {
    // Setup
    redis := setupTestRedis()
    ctx := context.Background()

    // Set test data
    redis.Set(ctx, "test_key", "test_value", time.Hour)

    // Test async operation
    result := redis.GetAsync(ctx, "test_key")

    // Wait with timeout
    value, err := result.WaitWithTimeout(5 * time.Second)

    assert.NoError(t, err)
    assert.Equal(t, "test_value", value)
}
```

### Integration Testing

```go
func TestAsyncDatabaseOperations(t *testing.T) {
    // Setup database
    db := setupTestDatabase()

    // Test batch operations
    users := []User{{Name: "Alice"}, {Name: "Bob"}}
    result := db.GORMCreateBatchAsync(ctx, users)

    _, errors := result.WaitAll()
    for _, err := range errors {
        assert.NoError(t, err)
    }
}
```

## Troubleshooting

### Common Issues

**Operations taking too long:**
- Check worker pool sizing
- Monitor queue depths
- Add timeouts to operations

**Memory leaks:**
- Ensure proper cleanup of resources
- Monitor goroutine counts
- Use context cancellation

**Deadlocks:**
- Avoid blocking operations in async jobs
- Use timeouts for all operations
- Monitor for circular dependencies

**Resource exhaustion:**
- Limit concurrent operations
- Implement backpressure
- Monitor system resources

## Conclusion

The async infrastructure implementation provides:

- **Non-blocking operations** for better application responsiveness
- **Controlled concurrency** through worker pools
- **Resource efficiency** with proper connection pooling
- **Fault tolerance** with timeout and error handling
- **Scalability** through batch operations and connection management

This approach ensures that your Go application can handle high loads while maintaining excellent user experience and system stability.
