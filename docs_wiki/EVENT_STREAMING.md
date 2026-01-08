# Live Event Streaming Documentation

## Overview

The Event Streaming System provides comprehensive real-time event streaming capabilities through a dual-implementation demonstration showcasing different architectural approaches:

### Service H - Event Streaming Showcase
**Dual-implementation demonstration** (`internal/services/modules/service_h.go`) showcasing both full implementation and utility approaches:
- **Full Implementation**: Complete event streaming with all broadcasting logic included
- **Utility Demo**: Clean implementation using `pkg/utils/broadcast.go`
- Two sets of API endpoints for comparison
- Educational example of different architectural patterns

### Broadcast Utility (`pkg/utils/broadcast.go`)
**Reusable broadcasting component** extracted for maximum reusability:
- Clean, well-documented utility for any service to use
- Thread-safe operations with proper synchronization
- Enhanced methods for monitoring and management
- Easy to integrate: just `utils.NewEventBroadcaster()`

## Service Comparison

| Implementation | Approach | API Prefix | Lines of Code | Benefits |
|----------------|----------|------------|---------------|----------|
| **Service H (Utility)** | Uses `pkg/utils/broadcast.go` | `/events/` | ~150 lines | Clean, simple, easy to understand |

Service H demonstrates how easy it is to implement event streaming using the broadcast utility.

## Key Features

- **Multiple Event Streams**: Support for concurrent event streams with independent client subscriptions
- **Server-Sent Events (SSE)**: Standards-compliant real-time push notifications
- **Event Broadcasting**: Send events to specific streams or broadcast to all streams
- **Automated Stream Generators**: Background processes generating sample events for demonstration
- **Stream Management**: Dynamic start, stop, pause, and resume operations for streams
- **Client Management**: Automatic subscription/unsubscription with buffered channels
- **Thread-Safe Operations**: Concurrent-safe event broadcasting and client management

## Architecture

### Core Components

#### 1. EventBroadcaster Utility (`pkg/utils/broadcast.go`)
The broadcast functionality has been extracted into a reusable utility in `pkg/utils/broadcast.go`. This makes it easy for any service to implement event streaming without duplicating code.

**Usage in Service G:**
```go
broadcaster := utils.NewEventBroadcaster()

// Subscribe to a stream
client := broadcaster.Subscribe("my-stream")

// Broadcast to a stream
broadcaster.Broadcast("my-stream", "event_type", "message", data)

// Broadcast to all streams
broadcaster.BroadcastToAll("global_event", "message", data)
```

**Core Utility Types:**
```go
type EventBroadcaster struct {
    streams   map[string][]*StreamClient // streamID -> clients
    clients   map[string]*StreamClient   // clientID -> client
    mu        sync.RWMutex
    nextID    int
    clientTTL time.Duration
}

type EventData struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`
    Message   string                 `json:"message"`
    Data      map[string]interface{} `json:"data,omitempty"`
    Timestamp int64                  `json:"timestamp"`
    StreamID  string                 `json:"stream_id,omitempty"`
}
```

#### 2. StreamClient
Represents a connected client for a specific event stream.

```go
type StreamClient struct {
    ID       string
    StreamID string
    Channel  chan EventData // Buffered channel (100 messages)
}
```

#### 3. EventData
Standardized event data structure sent to clients.

```go
type EventData struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`
    Message   string                 `json:"message"`
    Data      map[string]interface{} `json:"data,omitempty"`
    Timestamp int64                  `json:"timestamp"`
    StreamID  string                 `json:"stream_id,omitempty"`
}
```

#### 4. StreamGenerator
Manages automated event generation for demonstration streams.

```go
type StreamGenerator struct {
    streamID    string
    broadcaster *EventBroadcaster
    running     bool
    paused      bool
    stopChan    chan struct{}
    pauseChan   chan struct{}
    mu          sync.RWMutex
}
```

## API Endpoints

### Stream Subscription

#### GET `/api/v1/events/stream/{stream_id}`
Subscribe to a real-time event stream using Server-Sent Events.

**Parameters:**
- `stream_id` (path): The ID of the stream to subscribe to

**Response Headers:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
Access-Control-Allow-Origin: *
Access-Control-Allow-Headers: Cache-Control
```

**Event Format:**
```json
data: {"id":"evt_123456789","type":"user_action","message":"User logged in","data":{"user_id":"12345","action":"login"},"timestamp":1642598400,"stream_id":"default"}

data: {"id":"evt_123456790","type":"system_alert","message":"High CPU usage detected","data":{"cpu_percent":85.5,"threshold":80},"timestamp":1642598401,"stream_id":"system"}

```

**Example Usage (JavaScript):**
```javascript
const eventSource = new EventSource('/api/v1/events/stream/default');

eventSource.onmessage = function(event) {
    const eventData = JSON.parse(event.data);
    console.log('Received event:', eventData);

    switch(eventData.type) {
        case 'user_action':
            handleUserAction(eventData);
            break;
        case 'system_alert':
            handleSystemAlert(eventData);
            break;
        case 'notification':
            showNotification(eventData);
            break;
    }
};

eventSource.onerror = function(error) {
    console.error('EventSource failed:', error);
    eventSource.close();
};
```

**Example Usage (curl):**
```bash
curl -N -H "Accept: text/event-stream" \
     -H "Cache-Control: no-cache" \
     http://localhost:8080/api/v1/events/stream/default
```

### Event Broadcasting

#### POST `/api/v1/events/broadcast`
Broadcast an event to a specific stream or all streams.

**Request Body:**
```json
{
  "stream_id": "notifications",  // Optional: empty string broadcasts to all streams
  "type": "custom_event",
  "message": "Custom notification message",
  "data": {
    "priority": "high",
    "sender": "admin",
    "recipients": ["user1", "user2"]
  }
}
```

**Success Response:**
```json
{
  "success": true,
  "message": "Event broadcasted to stream: notifications",
  "timestamp": 1642598400
}
```

**Broadcast to All Streams:**
```json
{
  "type": "system_maintenance",
  "message": "System maintenance scheduled",
  "data": {
    "maintenance_window": "2024-01-15T02:00:00Z",
    "duration_minutes": 30
  }
}
```

### Stream Information

#### GET `/api/v1/events/streams`
Retrieve information about active streams and their connected clients.

**Response:**
```json
{
  "success": true,
  "data": {
    "default": {
      "clients": 5,
      "active": true,
      "generator": {
        "running": true,
        "paused": false
      }
    },
    "system": {
      "clients": 2,
      "active": true,
      "generator": {
        "running": true,
        "paused": false
      }
    },
    "notifications": {
      "clients": 0,
      "active": true,
      "generator": {
        "running": false,
        "paused": false
      }
    }
  },
  "timestamp": 1642598400
}
```

### Stream Management

#### POST `/api/v1/events/stream/{stream_id}/start`
Start or restart a stream generator.

**Response:**
```json
{
  "success": true,
  "message": "Stream 'system' started",
  "timestamp": 1642598400
}
```

#### POST `/api/v1/events/stream/{stream_id}/stop`
Stop a stream generator and remove it.

**Response:**
```json
{
  "success": true,
  "message": "Stream 'system' stopped and removed",
  "timestamp": 1642598400
}
```

#### POST `/api/v1/events/stream/{stream_id}/pause`
Pause a running stream generator.

**Response:**
```json
{
  "success": true,
  "message": "Stream 'system' paused",
  "timestamp": 1642598400
}
```

#### POST `/api/v1/events/stream/{stream_id}/resume`
Resume a paused stream generator.

**Response:**
```json
{
  "success": true,
  "message": "Stream 'system' resumed",
  "timestamp": 1642598400
}
```

## Default Streams

The service automatically starts four default streams with sample event generators:

### 1. `default` Stream
**Purpose:** General events and notifications
**Sample Events:**
- User actions (login, logout, profile updates)
- General system notifications
- Application-specific events

### 2. `system` Stream
**Purpose:** System-level alerts and metrics
**Sample Events:**
- High CPU/memory usage alerts
- Disk space warnings
- Service health status updates
- System performance metrics

### 3. `user-activity` Stream
**Purpose:** User action events
**Sample Events:**
- User authentication events
- Profile modifications
- Permission changes
- User-generated content updates

### 4. `notifications` Stream
**Purpose:** Application notifications
**Sample Events:**
- Push notifications
- Alert messages
- Scheduled reminders
- System announcements

## Event Types

### Predefined Event Types

- **`user_action`**: User-initiated actions (login, update profile, etc.)
- **`system_alert`**: System warnings and alerts (high CPU, low disk space)
- **`data_update`**: Database or data changes
- **`notification`**: General notifications and messages
- **`metric_update`**: System or application metrics
- **`stream_started`**: Stream initialization events
- **`connection`**: Client connection events

### Custom Event Types

You can define and use any custom event types in your applications:

```json
{
  "type": "order_created",
  "message": "New order received",
  "data": {
    "order_id": "ORD-12345",
    "customer_id": "CUST-67890",
    "total_amount": 299.99,
    "items": ["widget-a", "widget-b"]
  }
}
```

## Usage Examples

### Frontend Integration (React)

```javascript
import { useEffect, useState } from 'react';

function EventStreamComponent({ streamId }) {
    const [events, setEvents] = useState([]);
    const [isConnected, setIsConnected] = useState(false);

    useEffect(() => {
        const eventSource = new EventSource(`/api/v1/events/stream/${streamId}`);

        eventSource.onopen = () => {
            setIsConnected(true);
        };

        eventSource.onmessage = (event) => {
            const eventData = JSON.parse(event.data);
            setEvents(prev => [...prev.slice(-9), eventData]); // Keep last 10 events
        };

        eventSource.onerror = (error) => {
            console.error('EventSource error:', error);
            setIsConnected(false);
        };

        return () => {
            eventSource.close();
        };
    }, [streamId]);

    return (
        <div>
            <div>Connection Status: {isConnected ? '🟢 Connected' : '🔴 Disconnected'}</div>
            <div>
                {events.map((event, index) => (
                    <div key={index} className={`event event-${event.type}`}>
                        <strong>{event.type}:</strong> {event.message}
                        {event.data && <pre>{JSON.stringify(event.data, null, 2)}</pre>}
                    </div>
                ))}
            </div>
        </div>
    );
}
```

### Backend Event Broadcasting

```go
// Broadcast a custom event
func notifyOrderCreated(orderID string, customerID string, amount float64) {
    eventData := map[string]interface{}{
        "order_id": orderID,
        "customer_id": customerID,
        "amount": amount,
        "timestamp": time.Now().Unix(),
    }

    // This would be called from your event streaming service
    broadcastToStream("orders", "order_created", "New order received", eventData)
}

// Broadcast system alert
func alertHighCPU(cpuPercent float64) {
    eventData := map[string]interface{}{
        "cpu_percent": cpuPercent,
        "threshold": 80.0,
        "severity": "warning",
    }

    broadcastToAllStreams("system_alert", fmt.Sprintf("High CPU usage: %.1f%%", cpuPercent), eventData)
}
```

### Python Client

```python
import json
import requests
import sseclient

def stream_events(stream_id):
    """Stream events from a specific stream"""
    url = f"http://localhost:8080/api/v1/events/stream/{stream_id}"

    response = requests.get(url, stream=True, headers={
        'Accept': 'text/event-stream',
        'Cache-Control': 'no-cache'
    })

    client = sseclient.SSEClient(response)

    for event in client.events():
        event_data = json.loads(event.data)
        print(f"Event: {event_data['type']} - {event_data['message']}")

        # Handle different event types
        if event_data['type'] == 'user_action':
            handle_user_action(event_data)
        elif event_data['type'] == 'system_alert':
            handle_system_alert(event_data)

def broadcast_event(stream_id, event_type, message, data=None):
    """Broadcast an event to a stream"""
    url = "http://localhost:8080/api/v1/events/broadcast"

    payload = {
        "stream_id": stream_id,
        "type": event_type,
        "message": message,
        "data": data or {}
    }

    response = requests.post(url, json=payload)
    return response.json()
```

## Configuration

The event streaming service is configured via `config.yaml`:

```yaml
services:
  service_g: true  # Enable the event streaming service
```

No additional configuration is required. The service automatically starts with the default streams when enabled.

## Performance Considerations

### Client Connections
- Each client connection uses a buffered channel (100 messages)
- Connections are automatically cleaned up when clients disconnect
- No persistent storage of events (events are ephemeral)

### Stream Scalability
- Multiple streams can run concurrently
- Each stream generator runs in its own goroutine
- Event broadcasting is thread-safe using RWMutex

### Memory Usage
- Event channels are buffered to prevent blocking
- Automatic cleanup of disconnected clients
- Configurable client TTL (currently 24 hours)

## Error Handling

### Connection Errors
- SSE connections automatically reconnect on network failures
- Client disconnection is handled gracefully
- Channel buffering prevents message loss during temporary disconnections

### Stream Errors
- Stream generators include panic recovery
- Failed broadcasts don't affect other streams
- Error events can be broadcast to notify clients of issues

### Validation
- Event type and message validation
- Stream ID validation for subscriptions
- JSON serialization error handling

## Security Considerations

### Authentication
- SSE endpoints inherit authentication from the main application
- Consider implementing stream-specific authentication if needed

### CORS
- CORS headers are configured for cross-origin requests
- Adjust `Access-Control-Allow-Origin` for production deployments

### Rate Limiting
- Consider implementing rate limiting for broadcast endpoints
- Monitor for abuse of stream creation/management endpoints

## Monitoring and Debugging

### Active Streams Monitoring
```bash
# Check active streams and client counts
curl http://localhost:8080/api/v1/events/streams
```

### Stream Management
```bash
# Start a stream
curl -X POST http://localhost:8080/api/v1/events/stream/custom/start

# Stop a stream
curl -X POST http://localhost:8080/api/v1/events/stream/custom/stop

# Pause a stream
curl -X POST http://localhost:8080/api/v1/events/stream/custom/pause
```

### Testing Event Broadcasting
```bash
# Broadcast to specific stream
curl -X POST http://localhost:8080/api/v1/events/broadcast \
  -H "Content-Type: application/json" \
  -d '{
    "stream_id": "default",
    "type": "test_event",
    "message": "Test message",
    "data": {"test": true}
  }'

# Broadcast to all streams
curl -X POST http://localhost:8080/api/v1/events/broadcast \
  -H "Content-Type: application/json" \
  -d '{
    "type": "global_test",
    "message": "Global test message"
  }'
```

## Troubleshooting

### Common Issues

**Events not received by clients:**
- Check if the stream exists and is running
- Verify client connection (CORS, network issues)
- Check server logs for SSE header issues

**High memory usage:**
- Monitor client connections and disconnections
- Check for goroutine leaks in stream generators
- Verify channel buffer sizes are appropriate

**Stream not starting:**
- Ensure the service is enabled in config.yaml
- Check for errors in stream generator initialization
- Verify there are no naming conflicts with existing streams

### Debug Mode

Enable debug logging to troubleshoot issues:

```yaml
app:
  debug: true
```

This will provide detailed logs about:
- Client connections/disconnections
- Event broadcasting
- Stream generator lifecycle
- Error conditions

## Best Practices

### Client Implementation
1. **Handle Reconnection**: Implement automatic reconnection logic
2. **Event Filtering**: Filter events on the client side when possible
3. **Connection Limits**: Limit concurrent SSE connections per client
4. **Error Handling**: Gracefully handle connection failures

### Server Implementation
1. **Resource Management**: Monitor and limit concurrent connections
2. **Event Validation**: Validate event data before broadcasting
3. **Stream Lifecycle**: Properly manage stream creation and cleanup
4. **Performance Monitoring**: Track event throughput and latency

### Event Design
1. **Consistent Schema**: Use consistent event structures
2. **Meaningful Types**: Choose descriptive event type names
3. **Data Enrichment**: Include relevant context in event data
4. **Versioning**: Consider event versioning for API evolution

## Integration Examples

### Real-time Dashboard Updates
```javascript
// Update dashboard metrics in real-time
const metricStream = new EventSource('/api/v1/events/stream/system');

metricStream.onmessage = function(event) {
    const data = JSON.parse(event.data);
    if (data.type === 'metric_update') {
        updateDashboardMetrics(data.data);
    }
};
```

### Notification System
```javascript
// Handle real-time notifications
const notificationStream = new EventSource('/api/v1/events/stream/notifications');

notificationStream.onmessage = function(event) {
    const notification = JSON.parse(event.data);
    showNotificationToast(notification.message, notification.data);
};
```

### Collaborative Editing
```javascript
// Real-time collaborative document editing
const collabStream = new EventSource('/api/v1/events/stream/document_123');

collabStream.onmessage = function(event) {
    const change = JSON.parse(event.data);
    if (change.type === 'document_change') {
        applyRemoteChange(change.data);
    }
};
```

## Implementation Comparison

### Full Implementation vs Utility Approach

Service H demonstrates two different architectural approaches for the same event streaming functionality:

#### Full Implementation (Service H Full)
- **Code**: ~400 lines with complete broadcasting logic
- **Approach**: Self-contained with all `EventBroadcaster`, `EventData`, and stream management
- **API Prefix**: `/events-full/`
- **Benefits**: Full control, no external dependencies
- **Trade-offs**: Code duplication, higher maintenance burden

#### Utility Approach (Service H Utility)
- **Code**: ~200 lines (50% less than full implementation)
- **Approach**: Uses `pkg/utils/broadcast.go` for core functionality
- **API Prefix**: `/events-util/`
- **Benefits**: Clean code, centralized maintenance, reusable across services
- **Trade-offs**: Less control over implementation details

### Using the Broadcast Utility

Any service can easily implement event streaming using the broadcast utility:

```go
package modules

import (
    "test-go/pkg/utils"
    "test-go/pkg/logger"
    "test-go/pkg/response"
    "github.com/labstack/echo/v4"
)

type MyEventService struct {
    enabled     bool
    broadcaster *utils.EventBroadcaster
    logger      *logger.Logger
}

func NewMyEventService(enabled bool, logger *logger.Logger) *MyEventService {
    return &MyEventService{
        enabled:     enabled,
        broadcaster: utils.NewEventBroadcaster(),
        logger:      logger,
    }
}

func (s *MyEventService) streamEvents(c echo.Context) error {
    streamID := c.Param("stream_id")

    // Subscribe using utility
    client := s.broadcaster.Subscribe(streamID)
    defer s.broadcaster.Unsubscribe(client.ID)

    // Set up SSE headers
    c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
    c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
    c.Response().Header().Set(echo.HeaderConnection, "keep-alive")

    // Listen for events
    for {
        select {
        case event := <-client.Channel:
            // Handle event using utils.EventData
            // Send SSE response...
        case <-c.Request().Context().Done():
            return nil
        }
    }
}

func (s *MyEventService) broadcastEvent(c echo.Context) error {
    // Broadcast using utility
    s.broadcaster.Broadcast("my-stream", "custom_event", "My message", data)
    return response.Success(c, nil, "Event broadcasted")
}
```

### Utility Methods Reference

The broadcast utility provides these methods:

```go
// Core functionality
broadcaster := utils.NewEventBroadcaster()
client := broadcaster.Subscribe("stream_id")
broadcaster.Unsubscribe(clientID)
broadcaster.Broadcast(streamID, eventType, message, data)
broadcaster.BroadcastToAll(eventType, message, data)

// Monitoring and statistics
activeStreams := broadcaster.GetActiveStreams()      // map[string]int
streamClients := broadcaster.GetStreamClients(id)   // []*StreamClient
totalClients := broadcaster.GetTotalClients()       // int
streamCount := broadcaster.GetStreamCount()         // int
isActive := broadcaster.IsStreamActive(streamID)    // bool
```

### Benefits of the Utility Approach

1. **Code Reuse**: One implementation, multiple services
2. **Consistency**: All services use the same broadcasting logic
3. **Maintainability**: Bug fixes and improvements benefit all services
4. **Clean Services**: Services focus on business logic, not infrastructure
5. **Easy Testing**: Utility can be tested independently
6. **Performance**: Optimized, thread-safe implementation

### Migration Guide

To migrate an existing service to use the broadcast utility:

1. **Remove** duplicate `EventBroadcaster`, `EventData`, `StreamClient` types
2. **Import** `"test-go/pkg/utils"`
3. **Replace** `NewEventBroadcaster()` with `utils.NewEventBroadcaster()`
4. **Update** method calls to use `utils.EventData` type
5. **Test** the service functionality remains intact

### Testing Both Services

```bash
# Test Service G (full implementation)
curl -N http://localhost:8080/api/v1/events/stream/default

# Test Service H (utility demo)
curl -N http://localhost:8080/api/v1/broadcast-demo/stream/demo-notifications

# Check active streams for both services
curl http://localhost:8080/api/v1/events/streams
curl http://localhost:8080/api/v1/broadcast-demo/streams

# Broadcast events
curl -X POST http://localhost:8080/api/v1/events/broadcast \
  -H "Content-Type: application/json" \
  -d '{"type":"test","message":"Hello from Service G"}'

curl -X POST http://localhost:8080/api/v1/broadcast-demo/broadcast \
  -H "Content-Type: application/json" \
  -d '{"type":"test","message":"Hello from Service H"}'
```

This event streaming system provides both comprehensive full implementations and clean utility-based approaches, giving developers flexibility in how they implement real-time features while maintaining consistency and reusability.
