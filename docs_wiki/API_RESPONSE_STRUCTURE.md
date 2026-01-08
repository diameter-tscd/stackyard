# API Response Structure

## Overview

This document describes the standardized request/response structure for the Echo service.

## Response Format

All API responses follow a consistent structure using the `response.Response` type.

### Success Response

```json
{
  "success": true,
  "message": "Optional success message",
  "data": {
    // Your response data here
  },
  "meta": {
    "page": 1,
    "per_page": 10,
    "total": 100,
    "total_pages": 10
  },
  "timestamp": 1672531200
}
```

### Error Response

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {
      "field": "Additional error details"
    }
  },
  "timestamp": 1672531200
}
```

## Usage Examples

### Basic Success Response

```go
import (
    "test-go/pkg/response"
    "github.com/labstack/echo/v4"
)

func GetUser(c echo.Context) error {
    user := map[string]string{
        "id": "123",
        "name": "John Doe",
    }
    
    return response.Success(c, user, "User retrieved successfully")
}
```

### Paginated Response

```go
func GetUsers(c echo.Context) error {
    // Parse pagination from query
    var pagination response.PaginationRequest
    if err := c.Bind(&pagination); err != nil {
        return response.BadRequest(c, "Invalid pagination parameters")
    }
    
    // Get data (example)
    users := []User{} // Your users data
    total := int64(100)
    
    // Calculate meta
    meta := response.CalculateMeta(
        pagination.GetPage(),
        pagination.GetPerPage(),
        total,
    )
    
    return response.SuccessWithMeta(c, users, meta, "Users retrieved")
}
```

### Error Responses

```go
// Bad Request
if err := validate(data); err != nil {
    return response.BadRequest(c, "Invalid data", map[string]interface{}{
        "validation_errors": err,
    })
}

// Not Found
user := findUser(id)
if user == nil {
    return response.NotFound(c, "User not found")
}

// Unauthorized
if !isAuthenticated {
    return response.Unauthorized(c, "Invalid credentials")
}

// Internal Server Error
if err := processData(); err != nil {
    return response.InternalServerError(c, "Failed to process data")
}
```

### Request Validation

```go
import (
    "test-go/pkg/request"
    "test-go/pkg/response"
)

type CreateUserRequest struct {
    Username string `json:"username" validate:"required,username"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"required,gte=18,lte=100"`
    Phone    string `json:"phone" validate:"required,phone"`
}

func CreateUser(c echo.Context) error {
    var req CreateUserRequest
    
    // Bind and validate in one call
    if err := request.Bind(c, &req); err != nil {
        if validationErr, ok := err.(*request.ValidationError); ok {
            return response.ValidationError(c, "Validation failed", validationErr.GetFieldErrors())
        }
        return response.BadRequest(c, err.Error())
    }
    
    // Process the valid request
    user := createUser(req)
    
    return response.Created(c, user, "User created successfully")
}
```

## Available Response Helpers

### Success Responses
- `response.Success(c, data, message)` - 200 OK
- `response.SuccessWithMeta(c, data, meta, message)` - 200 OK with metadata
- `response.Created(c, data, message)` - 201 Created
- `response.NoContent(c)` - 204 No Content

### Error Responses
- `response.BadRequest(c, message, details)` - 400 Bad Request
- `response.Unauthorized(c, message)` - 401 Unauthorized
- `response.Forbidden(c, message)` - 403 Forbidden
- `response.NotFound(c, message)` - 404 Not Found
- `response.Conflict(c, message, details)` - 409 Conflict
- `response.ValidationError(c, message, details)` - 422 Unprocessable Entity
- `response.InternalServerError(c, message)` - 500 Internal Server Error
- `response.ServiceUnavailable(c, message)` - 503 Service Unavailable
- `response.Error(c, statusCode, errorCode, message, details)` - Custom error

## Pagination

The `PaginationRequest` struct provides convenient methods:

```go
type PaginationRequest struct {
    Page    int    `query:"page" json:"page"`
    PerPage int    `query:"per_page" json:"per_page"`
    Sort    string `query:"sort" json:"sort,omitempty"`
    Order   string `query:"order" json:"order,omitempty"`
}

// Methods
pagination.GetPage()      // Returns page (default: 1)
pagination.GetPerPage()   // Returns per_page (default: 10, max: 100)
pagination.GetOffset()    // Calculates offset for DB queries
pagination.GetOrder()     // Returns order (default: "desc")
```

## Request Validation

### Built-in Validators
- `required` - Field must not be empty
- `email` - Valid email format
- `min`, `max` - String length or numeric range
- `gte`, `lte` - Greater/less than or equal
- `oneof` - Value must be one of the specified options

### Custom Validators
- `phone` - Valid phone number format
- `username` - Alphanumeric username (3-20 chars)

### Common Request Structs

```go
// ID Request
type IDRequest struct {
    ID string `param:"id" validate:"required"`
}

// Search Request
type SearchRequest struct {
    Query  string            `query:"q" json:"query"`
    Filter map[string]string `query:"filter" json:"filter,omitempty"`
    Page   int               `query:"page" json:"page"`
    Limit  int               `query:"limit" json:"limit"`
}

// Date Range Request
type DateRangeRequest struct {
    StartDate string `query:"start_date" json:"start_date"`
    EndDate   string `query:"end_date" json:"end_date"`
}

// Sort Request
type SortRequest struct {
    SortBy    string `query:"sort_by" json:"sort_by"`
    SortOrder string `query:"sort_order" json:"sort_order"`
}
```

## Best Practices

1. **Always use the response helpers** - Don't manually construct response objects
2. **Include meaningful error codes** - Make errors machine-readable
3. **Provide context in error messages** - Help developers debug issues
4. **Use validation** - Validate all incoming requests
5. **Return appropriate status codes** - Follow HTTP standards
6. **Include timestamps** - All responses include Unix timestamps
7. **Use pagination** - For list endpoints, always support pagination
8. **Keep responses consistent** - All endpoints should follow the same structure

## Exposed Endpoints (Service A)
- `GET /api/v1/users` - List users
- `GET /api/v1/users/:id` - Get user details
- `POST /api/v1/users` - Create user
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

## Example Complete Handler

```go
package modules

import (
    "test-go/pkg/request"
    "test-go/pkg/response"
    "github.com/labstack/echo/v4"
)

type CreateTaskRequest struct {
    Title       string `json:"title" validate:"required,min=3,max=100"`
    Description string `json:"description" validate:"max=500"`
    Priority    string `json:"priority" validate:"required,oneof=low medium high"`
    DueDate     string `json:"due_date"`
}

type TaskResponse struct {
    ID          string `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description"`
    Priority    string `json:"priority"`
    Status      string `json:"status"`
    CreatedAt   int64  `json:"created_at"`
}

func CreateTask(c echo.Context) error {
    // Bind and validate
    var req CreateTaskRequest
    if err := request.Bind(c, &req); err != nil {
        if validationErr, ok := err.(*request.ValidationError); ok {
            return response.ValidationError(c, "Validation failed", validationErr.GetFieldErrors())
        }
        return response.BadRequest(c, err.Error())
    }
    
    // Business logic
    task, err := saveTask(req)
    if err != nil {
        return response.InternalServerError(c, "Failed to create task")
    }
    
    // Return success
    return response.Created(c, task, "Task created successfully")
}

func GetTasks(c echo.Context) error {
    // Parse pagination
    var pagination response.PaginationRequest
    c.Bind(&pagination)
    
    // Get tasks from database
    tasks, total, err := fetchTasks(pagination.GetOffset(), pagination.GetPerPage())
    if err != nil {
        return response.InternalServerError(c, "Failed to fetch tasks")
    }
    
    // Return with metadata
    meta := response.CalculateMeta(pagination.GetPage(), pagination.GetPerPage(), total)
    return response.SuccessWithMeta(c, tasks, meta)
}

func GetTask(c echo.Context) error {
    id := c.Param("id")
    
    task := findTask(id)
    if task == nil {
        return response.NotFound(c, "Task not found")
    }
    
    return response.Success(c, task)
}

func DeleteTask(c echo.Context) error {
    id := c.Param("id")
    
    if err := deleteTask(id); err != nil {
        return response.InternalServerError(c, "Failed to delete task")
    }
    
    return response.NoContent(c)
}
```
