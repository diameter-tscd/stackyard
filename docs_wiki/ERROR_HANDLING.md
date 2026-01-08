# HTTP Error Handling

## Overview

This document describes the custom HTTP error handling implementation that ensures all error responses are returned in a consistent JSON format rather than the default Echo HTML responses.

## Custom Error Handler

The server implements a custom HTTP error handler located in `internal/server/server.go`. This handler intercepts all HTTP errors and converts them to standardized JSON responses.

### Key Features

- All 404 (Not Found) errors return a specific JSON response with incident tracking
- All HTTP errors return JSON instead of HTML
- Non-HTTP errors are caught and return a 500 Internal Server Error
- Error details include correlation ID for debugging and tracking

## 404 Not Found Response

When a request is made to an endpoint that does not exist, the server returns the following JSON response:

### Response Format

```json
{
  "success": false,
  "status": 404,
  "error": {
    "code": "ENDPOINT_NOT_FOUND",
    "message": "Endpoint not found. This incident will be reported.",
    "details": {
      "path": "/api/v1/unknown-path",
      "method": "GET"
    }
  },
  "timestamp": 1734235788,
  "datetime": "2024-12-15T10:09:48+07:00",
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `success` | boolean | Always `false` for error responses |
| `status` | integer | HTTP status code (404) |
| `error.code` | string | Error code identifier (`ENDPOINT_NOT_FOUND`) |
| `error.message` | string | Human-readable error message |
| `error.details.path` | string | The requested URL path that was not found |
| `error.details.method` | string | The HTTP method used (GET, POST, etc.) |
| `timestamp` | integer | Unix timestamp of the response |
| `datetime` | string | ISO8601 formatted datetime |
| `correlation_id` | string | Unique request ID for tracking and debugging |

## Other HTTP Errors

For other HTTP errors (400, 401, 403, 405, 500, etc.), the server returns a simplified JSON response:

```json
{
  "success": false,
  "status": 405,
  "error": {
    "code": "HTTP_ERROR",
    "message": "Method Not Allowed"
  },
  "timestamp": 1734235788,
  "datetime": "2024-12-15T10:09:48+07:00",
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

## Internal Server Errors

For unexpected non-HTTP errors, the server returns a 500 Internal Server Error:

```json
{
  "success": false,
  "status": 500,
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "An unexpected error occurred"
  },
  "timestamp": 1734235788,
  "datetime": "2024-12-15T10:09:48+07:00",
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

## Error Codes Reference

| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| 400 | `BAD_REQUEST` | Invalid request parameters |
| 401 | `UNAUTHORIZED` | Authentication required |
| 403 | `FORBIDDEN` | Access denied |
| 404 | `ENDPOINT_NOT_FOUND` | Requested endpoint does not exist |
| 404 | `NOT_FOUND` | Requested resource not found (used in handlers) |
| 405 | `HTTP_ERROR` | Method not allowed |
| 409 | `CONFLICT` | Resource conflict |
| 422 | `VALIDATION_ERROR` | Request validation failed |
| 500 | `INTERNAL_ERROR` | Internal server error |
| 503 | `SERVICE_UNAVAILABLE` | Service temporarily unavailable |

## Implementation Details

### Location

The error handler is implemented in the `New()` function in `internal/server/server.go`:

```go
e.HTTPErrorHandler = func(err error, c echo.Context) {
    l.Error("HTTP Error", err)

    // Handle HTTP errors with JSON response
    if he, ok := err.(*echo.HTTPError); ok {
        var message string
        code := he.Code

        // Custom message for 404 Not Found
        if code == 404 {
            message = "Endpoint not found. This incident will be reported."
            response.Error(c, code, "ENDPOINT_NOT_FOUND", message, map[string]interface{}{
                "path":   c.Request().URL.Path,
                "method": c.Request().Method,
            })
            return
        }

        // For other HTTP errors, use the original message if it's a string
        if msg, ok := he.Message.(string); ok {
            message = msg
        } else {
            message = "An unexpected error occurred"
        }
        response.Error(c, code, "HTTP_ERROR", message)
        return
    }

    // For non-HTTP errors, return internal server error
    response.InternalServerError(c, "An unexpected error occurred")
}
```

### Logging

All HTTP errors are logged with the following information:

- Error message
- Stack trace (if available)
- Request context

Logs can be viewed in the monitoring interface when monitoring is enabled.

## Testing

### Test 404 Response

```bash
# Request to unknown endpoint
curl -X GET http://localhost:8080/api/v1/unknown-endpoint

# Expected response
{
  "success": false,
  "status": 404,
  "error": {
    "code": "ENDPOINT_NOT_FOUND",
    "message": "Endpoint not found. This incident will be reported.",
    "details": {
      "path": "/api/v1/unknown-endpoint",
      "method": "GET"
    }
  },
  "timestamp": 1734235788,
  "datetime": "2024-12-15T10:09:48+07:00",
  "correlation_id": "..."
}
```

### Test Method Not Allowed

```bash
# POST request to GET-only endpoint
curl -X POST http://localhost:8080/health

# Expected response
{
  "success": false,
  "status": 405,
  "error": {
    "code": "HTTP_ERROR",
    "message": "Method Not Allowed"
  },
  "timestamp": 1734235788,
  "datetime": "2024-12-15T10:09:48+07:00",
  "correlation_id": "..."
}
```

## Related Documentation

- [API Response Structure](./API_RESPONSE_STRUCTURE.md) - Detailed response format documentation
- [Integration Guide](./INTEGRATION_GUIDE.md) - How to handle errors in client applications
