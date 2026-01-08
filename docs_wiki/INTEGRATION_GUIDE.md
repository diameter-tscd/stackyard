# Infrastructure Integration Guide

This guide documents how to use the integrated infrastructure components in this boilerplate. All components are designed to be modular and can be enabled/disable via `config.yaml`.

## 1. Redis

### Configuration (`config.yaml`)
```yaml
redis:
  enabled: true
  host: "localhost"
  port: "6379"
  password: ""
  db: 0
```

### Usage (Code)
The `RedisManager` provides async operations with worker pools.

```go
// Inject RedisManager into your service
type MyService struct {
    redis *infrastructure.RedisManager
}

func (s *MyService) Example() {
    ctx := context.Background()

    // Async SET - returns immediately
    result := s.redis.SetAsync(ctx, "my-key", "my-value", time.Minute*10)

    // Wait for completion
    err := result.Wait()

    // Async GET
    getResult := s.redis.GetAsync(ctx, "my-key")
    value, err := getResult.Wait()

    // Async DELETE
    delResult := s.redis.DeleteAsync(ctx, "my-key")
    err := delResult.Wait()
}

// Batch operations for efficiency
func (s *MyService) BatchExample() {
    ctx := context.Background()
    keys := []string{"key1", "key2", "key3"}

    // Get multiple keys concurrently
    result := s.redis.GetBatchAsync(ctx, keys)
    values, errors := result.WaitAll()

    // Process results
    for i, val := range values {
        if errors[i] != nil {
            // Handle error
        } else {
            // Process value
        }
    }
}
```

---

## 2. Postgres

### Configuration (`config.yaml`)

The application now supports both single and multiple PostgreSQL connections for enhanced flexibility and scalability. Multiple connections allow you to monitor and query different databases through the web monitoring interface.

#### Single Connection (Original Format - Still Supported)
```yaml
postgres:
  enabled: true
  host: "localhost"
  port: "5432"
  user: "postgres"
  password: "password"
  dbname: "mydb"
  sslmode: "disable"
  max_open_conns: 10
  max_idle_conns: 5
```

#### Multiple Connections (New Format)
```yaml
postgres:
  enabled: true
  connections:
    - name: "primary"
      enabled: true
      host: "localhost"
      port: "5432"
      user: "postgres"
      password: "password"
      dbname: "primary_db"
      sslmode: "disable"

    - name: "secondary"
      enabled: true
      host: "localhost"
      port: "5433"
      user: "postgres"
      password: "password"
      dbname: "secondary_db"
      sslmode: "disable"

    - name: "analytics"
      enabled: false  # Disabled by default
      host: "analytics.example.com"
      port: "5432"
      user: "analytics_user"
      password: "analytics_password"
      dbname: "analytics_db"
      sslmode: "require"
```

### Usage (Code)

The application provides both single connection and multi-connection managers for PostgreSQL.

#### Using Single Connection (Backward Compatible)
```go
// Inject PostgresManager
type MyService struct {
    db *infrastructure.PostgresManager
}

func (s *MyService) Example() {
    // Access underlying sqlx.DB
    var users []User
    err := s.db.DB.Select(&users, "SELECT * FROM users WHERE active = $1", true)

    // Using transaction helper (if implemented in your manager extensions) or usage of standard sqlx patterns
    tx, err := s.db.DB.Beginx()
    // ...
}
```

#### Using Multiple Connections
```go
// Inject PostgresConnectionManager
type MyService struct {
    postgresManager *infrastructure.PostgresConnectionManager
}

func (s *MyService) Example() {
    // Get a specific named connection
    if conn, exists := s.postgresManager.GetConnection("primary"); exists {
        var users []User
        err := conn.DB.Select(&users, "SELECT * FROM users WHERE active = $1", true)
    }

    // Get the default connection (first enabled connection)
    if defaultConn, exists := s.postgresManager.GetDefaultConnection(); exists {
        // Use the default connection
    }

    // Get all connections
    allConnections := s.postgresManager.GetAllConnections()

    // Get status for all connections
    status := s.postgresManager.GetStatus()
}
```

### Connection Management Methods

The `PostgresConnectionManager` provides several useful methods:

```go
// Get a specific named connection
conn, exists := postgresManager.GetConnection("primary")

// Get the default connection (first enabled connection)
defaultConn, exists := postgresManager.GetDefaultConnection()

// Get all connections as a map
allConnections := postgresManager.GetAllConnections()

// Get status for all connections (useful for monitoring)
statusMap := postgresManager.GetStatus()

// Close all connections (for graceful shutdown)
err := postgresManager.CloseAll()
```

### Best Practices

1. **Connection Naming**: Use descriptive names like "primary", "secondary", "analytics", "read_replica"
2. **Error Handling**: Always check if a connection exists before using it
3. **Resource Management**: Close connections properly during application shutdown
4. **Configuration**: Disable unused connections to avoid unnecessary resource consumption
5. **Monitoring**: Use the status methods to monitor connection health in your monitoring system

### Migration Guide

To migrate from single to multiple connections:

1. **Update Configuration**: Convert your existing PostgreSQL config to the new format
2. **Update Services**: Modify services to use the connection manager when needed
3. **Test**: Verify all database operations work with the new connection manager
4. **Monitor**: Check the monitoring dashboard to see all PostgreSQL connections

The system automatically handles backward compatibility, so existing single-connection configurations will continue to work without modification.

---

## 3. Kafka

### Configuration (`config.yaml`)
```yaml
kafka:
  enabled: true
  brokers: ["localhost:9092"]
  topic: "my-topic"
  group_id: "my-group"
```

### Usage (Code)
The `KafkaManager` handles producing messages.

```go
// Inject KafkaManager
type MyService struct {
    kafka *infrastructure.KafkaManager
}

func (s *MyService) SendNotification() {
    // Publish a message
    err := s.kafka.Publish("notification-topic", []byte("Hello Kafka"))
    
    // Publish with Key (if supported by your specific implementation extension, default Publish typically sends value)
}
```

---

## 4. MinIO (Object Storage)

### Configuration (`config.yaml`)
```yaml
monitoring:
  minio:
    enabled: true
    endpoint: "localhost:9003"
    access_key_id: "minioadmin"
    secret_access_key: "minioadmin"
    use_ssl: false
    bucket_name: "main"
```

### Usage (Code)
The `MinIOManager` simplifies file uploads and URL retrieval.

```go
// Inject MinIOManager
type MyService struct {
    storage *infrastructure.MinIOManager
}

func (s *MyService) UploadAvatar(fileHeader *multipart.FileHeader) {
    file, _ := fileHeader.Open()
    defer file.Close()

    // Upload
    info, err := s.storage.UploadFile(context.Background(), "avatars/user-1.jpg", file, fileHeader.Size, "image/jpeg")

    // Get Presigned URL (for private buckets) or direct URL
    url := s.storage.GetFileUrl("avatars/user-1.jpg")
}
```

---

## 5. Cron Jobs

### Configuration (`config.yaml`)
Cron jobs can be defined in config for simple logging/testing, or registered in code for logic.

```yaml
cron:
  enabled: true
  jobs:
    "cleanup_logs": "0 0 * * *"   # Run at midnight
    "health_check": "*/5 * * * *" # Run every 5 minutes
```

### Usage (Code)
The `CronManager` allows dynamic job registration.

```go
// Inject CronManager
type MyService struct {
    cron *infrastructure.CronManager
}

func (s *MyService) InitJobs() {
    // Register a new job
    id, err := s.cron.AddJob("database_backup", "0 3 * * *", func() {
        fmt.Println("Performing database backup...")
        // Call service logic here
    })
}
```
