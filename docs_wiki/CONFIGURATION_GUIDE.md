# Configuration Guide

This document provides a comprehensive guide to configuring the application using the `config.yaml` file. It includes all available configuration options with explanations and examples.

## Table of Contents

- [Basic Configuration](#basic-configuration)
- [Server Configuration](#server-configuration)
- [Services Configuration](#services-configuration)
- [Authentication](#authentication)
- [Redis Configuration](#redis-configuration)
- [Kafka Configuration](#kafka-configuration)
- [PostgreSQL Configuration](#postgresql-configuration)
  - [Single Connection (Legacy)](#single-connection-legacy)
  - [Multiple Connections (Recommended)](#multiple-connections-recommended)
- [Monitoring Configuration](#monitoring-configuration)
  - [MinIO Configuration](#minio-configuration)
  - [External Services](#external-services)
- [Cron Jobs](#cron-jobs)
- [Encryption](#encryption)

## Basic Configuration

```yaml
app:
  name: "My Fancy Go App"
  version: "1.0.0"
  debug: true
  env: "development"
  banner_path: "banner.txt"
  startup_delay: 3        # seconds to display boot screen (0 to skip)
  quiet_startup: true     # suppress console logs (TUI only, logs still go to monitoring)
  enable_tui: true        # enable fancy TUI mode (false = traditional console logging)
```

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `name` | string | Application name | "My Fancy Go App" |
| `version` | string | Application version | "1.0.0" |
| `debug` | boolean | Enable debug mode | false |
| `env` | string | Environment (development, production, etc.) | "development" |
| `banner_path` | string | Path to banner text file | "banner.txt" |
| `startup_delay` | integer | Seconds to display boot screen (0 to skip) | 0 |
| `quiet_startup` | boolean | Suppress console logs during startup | false |
| `enable_tui` | boolean | Enable Terminal User Interface | false |

## Server Configuration

```yaml
server:
  port: "8080"
```

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `port` | string | Server port | "8080" |

## Services Configuration

```yaml
services:
  service_a: true
  service_b: false
  service_c: true
  service_d: false
  service_encryption: false
```

Each service can be enabled or disabled individually. Set to `true` to enable, `false` to disable.

## Authentication

```yaml
auth:
  type: "apikey"
  secret: "super-secret-key"
```

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `type` | string | Authentication type | "apikey" |
| `secret` | string | Secret key for authentication | "" |

## Redis Configuration

```yaml
redis:
  enabled: false
  address: "localhost:6379"
  password: ""
  db: 0
```

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `enabled` | boolean | Enable Redis | false |
| `address` | string | Redis server address | "localhost:6379" |
| `password` | string | Redis password | "" |
| `db` | integer | Redis database number | 0 |

## Kafka Configuration

```yaml
kafka:
  enabled: false
  brokers:
    - "localhost:9092"
  topic: "my-topic"
  group_id: "my-group"
```

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `enabled` | boolean | Enable Kafka | false |
| `brokers` | array | List of Kafka broker addresses | ["localhost:9092"] |
| `topic` | string | Default Kafka topic | "my-topic" |
| `group_id` | string | Consumer group ID | "my-group" |

## PostgreSQL Configuration

The application supports both single and multiple PostgreSQL connections. Multiple connections allow you to connect to different databases and switch between them dynamically through the web monitoring interface.

### Single Connection (Legacy)

```yaml
postgres:
  enabled: true
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "Mypostgres01"
  dbname: "primary_db"
  sslmode: "disable"
  max_open_conns: 10
  max_idle_conns: 5
```

### Multiple Connections (Recommended)

```yaml
postgres:
  enabled: true
  connections:
    - name: "primary"
      enabled: true
      host: "localhost"
      port: 5432
      user: "postgres"
      password: "Mypostgres01"
      dbname: "primary_db"
      sslmode: "disable"

    - name: "secondary"
      enabled: true
      host: "localhost"
      port: 5433
      user: "postgres"
      password: "Mypostgres01"
      dbname: "secondary_db"
      sslmode: "disable"

    - name: "analytics"
      enabled: false  # Disabled by default
      host: "analytics.example.com"
      port: 5432
      user: "analytics_user"
      password: "analytics_password"
      dbname: "analytics_db"
      sslmode: "require"
```

### Multiple Connections Features

- **Dynamic Switching**: Switch between database connections through the web monitoring interface
- **Connection Health**: Monitor the status of each database connection individually
- **Selective Queries**: Run queries on specific databases by selecting the connection
- **Load Distribution**: Distribute read/write operations across multiple databases
- **Failover Support**: Automatic fallback when connections become unavailable

### Web Monitoring Interface

When multiple connections are configured, the PostgreSQL monitoring page (`/monitoring/postgres`) provides:

- **Connection Selector**: Dropdown to choose which database to monitor/query
- **Status Indicators**: Green/red dots showing connection health for each database
- **Database Info**: Information about the currently selected database
- **Query Execution**: Run SQL queries on the selected database connection
- **Running Queries**: View active queries on the selected database

### Usage Examples

#### Monitoring Multiple Databases

1. Configure multiple connections in `config.yaml`
2. Start the application with monitoring enabled
3. Access the monitoring dashboard at `http://localhost:9090`
4. Navigate to the "Postgres" tab
5. Use the connection dropdown to switch between databases
6. Monitor health and run queries on each database individually

#### High Availability Setup

```yaml
postgres:
  enabled: true
  connections:
    - name: "primary"
      enabled: true
      host: "db-primary.example.com"
      port: 5432
      user: "app_user"
      password: "${PRIMARY_DB_PASSWORD}"
      dbname: "app_db"
      sslmode: "require"

    - name: "replica"
      enabled: true
      host: "db-replica.example.com"
      port: 5432
      user: "app_user"
      password: "${REPLICA_DB_PASSWORD}"
      dbname: "app_db"
      sslmode: "require"
```

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `enabled` | boolean | Enable PostgreSQL | false |
| `connections` | array | List of PostgreSQL connections | [] |
| `connections[].name` | string | Connection name identifier | "" |
| `connections[].enabled` | boolean | Enable this specific connection | false |
| `connections[].host` | string | PostgreSQL host | "localhost" |
| `connections[].port` | integer | PostgreSQL port | 5432 |
| `connections[].user` | string | PostgreSQL username | "postgres" |
| `connections[].password` | string | PostgreSQL password | "" |
| `connections[].dbname` | string | Database name | "" |
| `connections[].sslmode` | string | SSL mode (disable, require, etc.) | "disable" |

## Monitoring Configuration

```yaml
monitoring:
  enabled: true
  port: "9090"
  password: "admin"
  obfuscate_api: true
  title: "GoBP Admin"
  subtitle: "My Kisah Emuach ❤️"
  max_photo_size_mb: 2
  upload_dir: "web/monitoring/uploads"
```

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `enabled` | boolean | Enable monitoring | false |
| `port` | string | Monitoring port | "9090" |
| `password` | string | Monitoring password | "" |
| `obfuscate_api` | boolean | Enable API obfuscation | false |
| `title` | string | Monitoring dashboard title | "GoBP Admin" |
| `subtitle` | string | Monitoring dashboard subtitle | "" |
| `max_photo_size_mb` | integer | Maximum photo upload size in MB | 2 |
| `upload_dir` | string | Upload directory path | "web/monitoring/uploads" |

### MinIO Configuration

```yaml
monitoring:
  minio:
    enabled: true
    endpoint: "localhost:9003"
    access_key: "minioadmin"
    secret_key: "minioadmin"
    use_ssl: false
    bucket: "main"
```

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `enabled` | boolean | Enable MinIO integration | false |
| `endpoint` | string | MinIO endpoint | "localhost:9000" |
| `access_key` | string | MinIO access key | "minioadmin" |
| `secret_key` | string | MinIO secret key | "minioadmin" |
| `use_ssl` | boolean | Use SSL for MinIO connection | false |
| `bucket` | string | Default bucket name | "main" |

### External Services

```yaml
monitoring:
  external:
    services:
      - name: "Google"
        url: "https://google.com"
      - name: "Soundcloud"
        url: "https://soundcloud.com"
      - name: "Local API"
        url: "http://localhost:8080/health"
```

## Cron Jobs

```yaml
cron:
  enabled: true
  jobs:
    log_cleanup: "0 0 * * *"          # Run at midnight
    health_check: "*/10 * * * * *"    # Every 10 seconds
```

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `enabled` | boolean | Enable cron jobs | false |
| `jobs` | object | Cron job definitions | {} |

## Encryption

```yaml
encryption:
  enabled: false
  algorithm: "aes-256-gcm"
  key: ""
  rotate_keys: false
  key_rotation_interval: "24h"
```

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `enabled` | boolean | Enable encryption | false |
| `algorithm` | string | Encryption algorithm | "aes-256-gcm" |
| `key` | string | Encryption key (32 bytes for AES-256) | "" |
| `rotate_keys` | boolean | Enable automatic key rotation | false |
| `key_rotation_interval` | string | Key rotation interval | "24h" |

## Complete Example Configuration

```yaml
# Example configuration with multiple PostgreSQL connections
app:
  name: "My Fancy Go App"
  version: "1.0.0"
  debug: true
  env: "development"
  banner_path: "banner.txt"
  startup_delay: 3        # seconds to display boot screen (0 to skip)
  quiet_startup: true     # suppress console logs (TUI only, logs still go to monitoring)
  enable_tui: true        # enable fancy TUI mode (false = traditional console logging)

server:
  port: "8080"

services:
  service_a: true
  service_b: false
  service_c: true
  service_d: false
  service_encryption: false

auth:
  type: "apikey"
  secret: "super-secret-key"

redis:
  enabled: false
  address: "localhost:6379"
  password: ""
  db: 0

kafka:
  enabled: false
  brokers:
    - "localhost:9092"
  topic: "my-topic"
  group_id: "my-group"

# NEW: Multiple PostgreSQL connections configuration
postgres:
  enabled: true
  connections:
    - name: "primary"
      enabled: true
      host: "localhost"
      port: 5432
      user: "postgres"
      password: "Mypostgres01"
      dbname: "primary_db"
      sslmode: "disable"

    - name: "secondary"
      enabled: true
      host: "localhost"
      port: 5433
      user: "postgres"
      password: "Mypostgres01"
      dbname: "secondary_db"
      sslmode: "disable"

    - name: "analytics"
      enabled: false  # Disabled by default
      host: "analytics.example.com"
      port: 5432
      user: "analytics_user"
      password: "analytics_password"
      dbname: "analytics_db"
      sslmode: "require"

monitoring:
  enabled: true
  port: "9090"
  password: "admin"
  obfuscate_api: true
  title: "GoBP Admin"
  subtitle: "My Kisah Emuach ❤️"
  max_photo_size_mb: 2
  upload_dir: "web/monitoring/uploads"

  minio:
    enabled: true
    endpoint: "localhost:9003"
    access_key: "minioadmin"
    secret_key: "minioadmin"
    use_ssl: false
    bucket: "main"

  external:
    services:
      - name: "Google"
        url: "https://google.com"
      - name: "Soundcloud"
        url: "https://soundcloud.com"
      - name: "Local API"
        url: "http://localhost:8080/health"

cron:
  enabled: true
  jobs:
    log_cleanup: "0 0 * * *"
    health_check: "*/10 * * * * *"

encryption:
  enabled: false
  algorithm: "aes-256-gcm"
  key: ""
  rotate_keys: false
  key_rotation_interval: "24h"
```

## Usage

1. Copy the example configuration to `config.yaml`
2. Modify the values according to your environment
3. Ensure sensitive information (passwords, secrets) are properly secured
4. For production environments, consider using environment variables for sensitive data

## Best Practices

- Use environment variables for sensitive configuration (passwords, API keys)
- Disable unused services to reduce resource consumption
- Use meaningful names for PostgreSQL connections
- Monitor connection health through the monitoring dashboard
- Regularly rotate encryption keys if encryption is enabled
