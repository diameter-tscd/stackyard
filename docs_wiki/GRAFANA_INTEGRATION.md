# Grafana Integration Guide

This document provides comprehensive information about Grafana integration in the project, allowing users to easily integrate their applications with Grafana for monitoring and visualization.

## Overview

Grafana integration enables seamless dashboard creation, data source management, and real-time monitoring through a comprehensive API. The integration provides:

- **Dashboard Management**: Create, update, retrieve, and delete Grafana dashboards programmatically
- **Data Source Integration**: Configure and manage data sources for metrics collection
- **Annotation Support**: Add annotations to dashboards for event tracking
- **Health Monitoring**: Real-time health checks and status monitoring
- **Async Operations**: Non-blocking operations with worker pools for performance

## Features

- **Complete Dashboard API**: Full CRUD operations for Grafana dashboards
- **Data Source Management**: Create and configure data sources programmatically
- **Annotation System**: Add timeline annotations for events and incidents
- **Health Monitoring**: Real-time Grafana instance health checks
- **Async Operations**: Non-blocking API calls with proper error handling
- **Retry Logic**: Built-in retry mechanisms for reliable API communication
- **Type Safety**: Strongly typed structures for all Grafana entities

## Configuration

### Basic Configuration

Add Grafana configuration to your `config.yaml`:

```yaml
grafana:
  enabled: true
  url: "http://localhost:3000"
  api_key: "your-grafana-api-key"
  username: "admin"          # Optional, for basic auth
  password: "admin"          # Optional, for basic auth
```

### Configuration Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `enabled` | boolean | Yes | Enable/disable Grafana integration |
| `url` | string | Yes | Grafana server URL (e.g., `http://localhost:3000`) |
| `api_key` | string | No* | Grafana API key for authentication |
| `username` | string | No | Username for basic authentication |
| `password` | string | No | Password for basic authentication |

*Either `api_key` OR `username`/`password` must be provided for authentication.

### Service Configuration

Enable the Grafana service in your configuration:

```yaml
services:
  service_i: true
```

## API Endpoints

The Grafana integration service provides RESTful endpoints under `/api/v1/grafana`:

### Dashboard Management

#### Create Dashboard
```http
POST /api/v1/grafana/dashboards
Content-Type: application/json

{
  "title": "System Metrics",
  "tags": ["system", "metrics"],
  "timezone": "UTC",
  "panels": [
    {
      "id": 1,
      "title": "CPU Usage",
      "type": "graph",
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 0,
        "y": 0
      },
      "targets": [
        {
          "expr": "100 - (avg by(instance) (irate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100)",
          "legendFormat": "{{instance}}"
        }
      ]
    }
  ],
  "time": {
    "from": "now-1h",
    "to": "now"
  },
  "refresh": "5s"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Dashboard created successfully",
  "data": {
    "id": 123,
    "uid": "abc123def",
    "title": "System Metrics",
    "version": 1
  },
  "timestamp": 1642598400
}
```

#### Update Dashboard
```http
PUT /api/v1/grafana/dashboards/{uid}
Content-Type: application/json

{
  "title": "Updated System Metrics",
  "panels": [
    {
      "id": 1,
      "title": "Memory Usage",
      "type": "graph",
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 0,
        "y": 0
      },
      "targets": [
        {
          "expr": "100 - ((node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes) * 100)",
          "legendFormat": "Memory Usage %"
        }
      ]
    }
  ]
}
```

#### Get Dashboard
```http
GET /api/v1/grafana/dashboards/{uid}
```

**Response:**
```json
{
  "success": true,
  "message": "Dashboard retrieved successfully",
  "data": {
    "id": 123,
    "uid": "abc123def",
    "title": "System Metrics",
    "panels": [...],
    "time": {...},
    "version": 1
  },
  "timestamp": 1642598400
}
```

#### Delete Dashboard
```http
DELETE /api/v1/grafana/dashboards/{uid}
```

**Response:**
```json
{
  "success": true,
  "message": "Dashboard deleted successfully",
  "timestamp": 1642598400
}
```

#### List Dashboards
```http
GET /api/v1/grafana/dashboards?page=1&per_page=50
```

**Response:**
```json
{
  "success": true,
  "message": "Dashboards retrieved successfully",
  "data": [
    {
      "id": 123,
      "uid": "abc123def",
      "title": "System Metrics",
      "tags": ["system", "metrics"]
    },
    {
      "id": 124,
      "uid": "def456ghi",
      "title": "Application Metrics",
      "tags": ["app", "performance"]
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 50,
    "total": 2,
    "total_pages": 1
  },
  "timestamp": 1642598400
}
```

### Data Source Management

#### Create Data Source
```http
POST /api/v1/grafana/datasources
Content-Type: application/json

{
  "name": "Prometheus",
  "type": "prometheus",
  "url": "http://prometheus:9090",
  "access": "proxy",
  "basicAuth": false,
  "jsonData": {
    "timeInterval": "15s",
    "queryTimeout": "60s"
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "Data source created successfully",
  "data": {
    "id": 1,
    "uid": "prometheus-uid",
    "name": "Prometheus",
    "type": "prometheus",
    "url": "http://prometheus:9090"
  },
  "timestamp": 1642598400
}
```

### Annotation Management

#### Create Annotation
```http
POST /api/v1/grafana/annotations
Content-Type: application/json

{
  "dashboardId": 123,
  "panelId": 1,
  "time": 1642598400000,
  "timeEnd": 1642598460000,
  "tags": ["deployment", "v1.2.0"],
  "text": "Application deployed to production"
}
```

### Health Monitoring

#### Get Grafana Health
```http
GET /api/v1/grafana/health
```

**Response:**
```json
{
  "success": true,
  "message": "Grafana health check successful",
  "data": {
    "version": "9.3.0",
    "database": "ok",
    "commit": "abc123def"
  },
  "timestamp": 1642598400
}
```

## Usage Examples

### Creating a System Monitoring Dashboard

```bash
# Create a comprehensive system monitoring dashboard
curl -X POST http://localhost:8080/api/v1/grafana/dashboards \
  -H "Content-Type: application/json" \
  -d '{
    "title": "System Overview",
    "tags": ["system", "overview"],
    "panels": [
      {
        "id": 1,
        "title": "CPU Usage",
        "type": "graph",
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0},
        "targets": [{
          "expr": "100 - (avg by(instance) (irate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100)",
          "legendFormat": "CPU Usage %"
        }]
      },
      {
        "id": 2,
        "title": "Memory Usage",
        "type": "graph",
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0},
        "targets": [{
          "expr": "100 - ((node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes) * 100)",
          "legendFormat": "Memory Usage %"
        }]
      }
    ],
    "time": {"from": "now-1h", "to": "now"},
    "refresh": "30s"
  }'
```

### Setting Up Prometheus Data Source

```bash
# Configure Prometheus as a data source
curl -X POST http://localhost:8080/api/v1/grafana/datasources \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Prometheus",
    "type": "prometheus",
    "url": "http://prometheus:9090",
    "access": "proxy",
    "jsonData": {
      "timeInterval": "15s",
      "queryTimeout": "60s",
      "httpMethod": "POST"
    }
  }'
```

### Adding Deployment Annotations

```bash
# Add an annotation for a deployment event
curl -X POST http://localhost:8080/api/v1/grafana/annotations \
  -H "Content-Type: application/json" \
  -d '{
    "time": 1642598400000,
    "tags": ["deployment", "api", "v2.1.0"],
    "text": "API service deployed to production environment"
  }'
```

## Programmatic Usage

### Go Client Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "test-go/pkg/infrastructure"
    "test-go/config"
)

func main() {
    // Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatal(err)
    }

    // Create Grafana manager
    grafanaMgr, err := infrastructure.NewGrafanaManager(cfg.Grafana)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Create a simple dashboard
    dashboard := infrastructure.GrafanaDashboard{
        Title: "Application Metrics",
        Tags:  []string{"app", "metrics"},
        Panels: []infrastructure.GrafanaPanel{
            {
                ID:    1,
                Title: "Request Rate",
                Type:  "graph",
                GridPos: infrastructure.GrafanaGridPos{
                    H: 8,
                    W: 12,
                    X: 0,
                    Y: 0,
                },
                Targets: []infrastructure.GrafanaTarget{
                    {
                        Expr:         "rate(http_requests_total[5m])",
                        LegendFormat: "Request Rate",
                    },
                },
            },
        },
        Time: infrastructure.GrafanaTimeRange{
            From: "now-1h",
            To:   "now",
        },
        Refresh: "30s",
    }

    // Create dashboard asynchronously
    result := grafanaMgr.CreateDashboardAsync(ctx, dashboard)
    createdDashboard, err := result.Wait()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Dashboard created: %s (UID: %s)\n", createdDashboard.Title, createdDashboard.UID)
}
```

### Creating Data Sources Programmatically

```go
// Create Prometheus data source
prometheusDS := infrastructure.GrafanaDataSource{
    Name:   "Prometheus",
    Type:   "prometheus",
    URL:    "http://prometheus:9090",
    Access: "proxy",
    JSONData: map[string]interface{}{
        "timeInterval": "15s",
        "queryTimeout": "60s",
    },
}

result := grafanaMgr.CreateDataSourceAsync(ctx, prometheusDS)
createdDS, err := result.Wait()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Data source created: %s (ID: %d)\n", createdDS.Name, createdDS.ID)
```

## Integration Patterns

### Application Monitoring Dashboard

Create a comprehensive monitoring dashboard for your application:

```go
func createApplicationDashboard(grafanaMgr *infrastructure.GrafanaManager) error {
    dashboard := infrastructure.GrafanaDashboard{
        Title: "Application Monitoring",
        Tags:  []string{"app", "monitoring", "production"},
        Panels: []infrastructure.GrafanaPanel{
            // HTTP Request Rate
            {
                ID:    1,
                Title: "HTTP Request Rate",
                Type:  "graph",
                GridPos: infrastructure.GrafanaGridPos{H: 8, W: 12, X: 0, Y: 0},
                Targets: []infrastructure.GrafanaTarget{
                    {
                        Expr:         "rate(http_requests_total[5m])",
                        LegendFormat: "Requests/sec",
                    },
                },
            },
            // Error Rate
            {
                ID:    2,
                Title: "Error Rate",
                Type:  "graph",
                GridPos: infrastructure.GrafanaGridPos{H: 8, W: 12, X: 12, Y: 0},
                Targets: []infrastructure.GrafanaTarget{
                    {
                        Expr:         "rate(http_requests_total{status=~\"5..\"}[5m])",
                        LegendFormat: "Errors/sec",
                    },
                },
            },
            // Response Time
            {
                ID:    3,
                Title: "Response Time",
                Type:  "graph",
                GridPos: infrastructure.GrafanaGridPos{H: 8, W: 12, X: 0, Y: 8},
                Targets: []infrastructure.GrafanaTarget{
                    {
                        Expr:         "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
                        LegendFormat: "95th percentile",
                    },
                },
            },
        },
        Time: infrastructure.GrafanaTimeRange{
            From: "now-6h",
            To:   "now",
        },
        Refresh: "30s",
    }

    _, err := grafanaMgr.CreateDashboard(context.Background(), dashboard)
    return err
}
```

### Automated Alert Annotations

Automatically add annotations when alerts are triggered:

```go
func addAlertAnnotation(grafanaMgr *infrastructure.GrafanaManager, alert Alert) error {
    annotation := infrastructure.GrafanaAnnotation{
        Time:   alert.Timestamp,
        TimeEnd: alert.Timestamp + 300000, // 5 minutes
        Tags:   []string{"alert", alert.Severity, alert.Service},
        Text:   fmt.Sprintf("Alert: %s - %s", alert.Title, alert.Description),
        Data: map[string]interface{}{
            "severity": alert.Severity,
            "service":  alert.Service,
            "value":    alert.Value,
            "threshold": alert.Threshold,
        },
    }

    _, err := grafanaMgr.CreateAnnotation(context.Background(), annotation)
    return err
}
```

## Dashboard Templates

### System Metrics Dashboard

```json
{
  "title": "System Metrics",
  "tags": ["system", "infrastructure"],
  "panels": [
    {
      "id": 1,
      "title": "CPU Usage",
      "type": "graph",
      "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0},
      "targets": [{
        "expr": "100 - (avg by(instance) (irate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100)",
        "legendFormat": "{{instance}} CPU Usage %"
      }]
    },
    {
      "id": 2,
      "title": "Memory Usage",
      "type": "graph",
      "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0},
      "targets": [{
        "expr": "100 - ((node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes) * 100)",
        "legendFormat": "{{instance}} Memory Usage %"
      }]
    },
    {
      "id": 3,
      "title": "Disk Usage",
      "type": "graph",
      "gridPos": {"h": 8, "w": 12, "x": 0, "y": 8},
      "targets": [{
        "expr": "100 - ((node_filesystem_avail_bytes / node_filesystem_size_bytes) * 100)",
        "legendFormat": "{{instance}} {{mountpoint}}"
      }]
    }
  ],
  "time": {"from": "now-1h", "to": "now"},
  "refresh": "30s"
}
```

### Application Performance Dashboard

```json
{
  "title": "Application Performance",
  "tags": ["application", "performance"],
  "panels": [
    {
      "id": 1,
      "title": "Request Rate",
      "type": "graph",
      "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0},
      "targets": [{
        "expr": "rate(http_requests_total[5m])",
        "legendFormat": "Requests/sec"
      }]
    },
    {
      "id": 2,
      "title": "Response Time (95th percentile)",
      "type": "graph",
      "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0},
      "targets": [{
        "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
        "legendFormat": "95th percentile"
      }]
    },
    {
      "id": 3,
      "title": "Error Rate",
      "type": "graph",
      "gridPos": {"h": 8, "w": 12, "x": 0, "y": 8},
      "targets": [{
        "expr": "rate(http_requests_total{status=~\"5..\"}[5m]) / rate(http_requests_total[5m]) * 100",
        "legendFormat": "Error Rate %"
      }]
    }
  ],
  "time": {"from": "now-6h", "to": "now"},
  "refresh": "30s"
}
```

## Security Considerations

### Authentication
- **API Keys**: Use Grafana API keys with minimal required permissions
- **Basic Auth**: Avoid basic authentication in production environments
- **Network Security**: Ensure Grafana API is accessible only from trusted networks

### Permissions
- **Service Accounts**: Create dedicated service accounts for API access
- **Least Privilege**: Grant only necessary permissions for dashboard and data source management
- **Token Rotation**: Regularly rotate API keys and tokens

### Data Protection
- **Sensitive Data**: Avoid storing sensitive information in dashboard configurations
- **Access Control**: Implement proper access controls for dashboard visibility
- **Audit Logging**: Enable audit logging for dashboard changes

## Performance Optimization

### Async Operations
All Grafana operations support async execution:

```go
// Fire-and-forget dashboard creation
result := grafanaMgr.CreateDashboardAsync(ctx, dashboard)
// Continue with other operations while dashboard is being created
```

### Batch Operations
For multiple operations, consider batching:

```go
// Create multiple dashboards concurrently
dashboards := []infrastructure.GrafanaDashboard{dash1, dash2, dash3}
for _, dash := range dashboards {
    go func(d infrastructure.GrafanaDashboard) {
        _, err := grafanaMgr.CreateDashboard(ctx, d)
        if err != nil {
            log.Printf("Failed to create dashboard %s: %v", d.Title, err)
        }
    }(dash)
}
```

### Connection Pooling
The Grafana manager uses HTTP client with connection pooling for optimal performance.

## Error Handling

### Common Error Scenarios

1. **Connection Failed**
   ```
   Error: failed to connect to Grafana: Get "http://localhost:3000/api/health": dial tcp [::1]:3000: connect: connection refused
   Solution: Ensure Grafana is running and accessible
   ```

2. **Authentication Failed**
   ```
   Error: failed to create dashboard: 401 Unauthorized
   Solution: Verify API key or credentials are correct
   ```

3. **Invalid Dashboard JSON**
   ```
   Error: failed to create dashboard: 400 Bad Request
   Solution: Validate dashboard JSON structure
   ```

4. **Dashboard Not Found**
   ```
   Error: dashboard not found: abc123def
   Solution: Verify dashboard UID exists
   ```

### Retry Logic

The integration includes automatic retry logic for transient failures:

- **Max Retries**: 3 attempts
- **Backoff**: Exponential backoff (1s, 2s, 4s)
- **Timeout**: 30-second request timeout

## Monitoring and Observability

### Health Checks

Monitor Grafana integration health:

```bash
# Check Grafana service health
curl http://localhost:8080/api/v1/grafana/health

# Check overall infrastructure health
curl http://localhost:8080/health/infrastructure
```

### Metrics Integration

The Grafana integration can be monitored through the existing monitoring dashboard at `http://localhost:9090`.

### Logging

All Grafana operations are logged with structured logging:

```
INFO: Grafana dashboard created: System Metrics (UID: abc123def)
ERROR: Failed to create Grafana dashboard: 401 Unauthorized
```

## Troubleshooting

### Debug Mode

Enable debug logging for detailed operation information:

```yaml
app:
  debug: true
```

### Connection Issues

1. **Verify Grafana URL**: Ensure the configured URL is correct and accessible
2. **Check Authentication**: Validate API key or credentials
3. **Network Connectivity**: Test network connectivity to Grafana server
4. **Firewall Rules**: Ensure required ports are open

### API Issues

1. **Validate JSON**: Ensure dashboard and data source JSON is valid
2. **Check Permissions**: Verify API key has necessary permissions
3. **Rate Limits**: Check for API rate limiting issues
4. **Version Compatibility**: Ensure Grafana version supports requested features

### Performance Issues

1. **Async Operations**: Use async methods for non-blocking operations
2. **Batch Operations**: Group multiple operations to reduce API calls
3. **Connection Reuse**: HTTP client reuses connections automatically
4. **Timeout Configuration**: Adjust timeouts based on network conditions

## Integration Examples

### Docker Compose Setup

```yaml
version: '3.8'
services:
  app:
    # Your application configuration
    environment:
      - GRAFANA_ENABLED=true
      - GRAFANA_URL=http://grafana:3000
      - GRAFANA_API_KEY=your-api-key

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana_data:/var/lib/grafana

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml

volumes:
  grafana_data:
```

### Kubernetes Deployment

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  config.yaml: |
    grafana:
      enabled: true
      url: "http://grafana.grafana.svc.cluster.local:3000"
      api_key: "${GRAFANA_API_KEY}"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  template:
    spec:
      containers:
      - name: app
        env:
        - name: GRAFANA_API_KEY
          valueFrom:
            secretKeyRef:
              name: grafana-secrets
              key: api-key
```

## Conclusion

The Grafana integration provides a comprehensive, production-ready solution for programmatic dashboard and data source management. With support for async operations, retry logic, and comprehensive error handling, it enables seamless integration between your application and Grafana for monitoring and visualization needs.

The integration follows the same patterns as other infrastructure components in the project, ensuring consistency and maintainability. Whether you need to create dashboards automatically, manage data sources, or add annotations for events, the Grafana integration provides the tools and APIs to accomplish these tasks efficiently and reliably.
