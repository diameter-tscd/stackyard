# Default Request Response Structure for Echo Service

## Summary

This project now has a structured, clean, and dynamic request/response structure for the Echo service. This system provides:

**Standardized Response** - Consistent format for all API endpoints  
**Automatic Request Validation** - Input validation with clear error messages  
**Built-in Pagination** - Pagination support with complete metadata  
**Comprehensive Error Handling** - Various helper functions for error responses  
**Type-safe** - Uses structs for request/response  

---

## Created Files

### 1. **pkg/response/response.go**
Package for standardizing API responses:

**Structs:**
- `Response` - Main response structure
- `ErrorDetail` - Detailed error information
- `Meta` - Pagination metadata
- `PaginationRequest` - Standard pagination parameters

**Helper Functions:**
- Success responses: `Success()`, `SuccessWithMeta()`, `Created()`, `NoContent()`
- Error responses: `BadRequest()`, `Unauthorized()`, `Forbidden()`, `NotFound()`, `Conflict()`, `ValidationError()`, `InternalServerError()`, `ServiceUnavailable()`
- Utilities: `CalculateMeta()` for pagination metadata

### 2. **pkg/request/request.go**
Package for request validation and binding:

**Functions:**
- `Bind()` - Bind and validate request simultaneously
- `Validate()` - Validate struct using validator tags
- `FormatValidationErrors()` - Format error messages in user-friendly way

**Custom Validators:**
- `phone` - Phone number format validation
- `username` - Username validation (alphanumeric, 3-20 chars)

**Common Request Structs:**
- `IDRequest` - For requests with single ID
- `IDsRequest` - For requests with multiple IDs
- `SearchRequest` - For search with filter and pagination
- `DateRangeRequest` - For date-based filtering
- `SortRequest` - For sorting parameters

### 3. **docs/API_RESPONSE_STRUCTURE.md**
Complete documentation with:
- Response structure format
- Examples of all helper functions
- Best practices
- Complete handler examples

### 4. **docs/examples/response_examples.go**
Example implementation file:
- 7 different use cases
- Success, error, pagination, validation
- Search, custom errors, delete operations

### 5. **internal/services/modules/service_a.go** (Updated)
Updated as reference implementation with:
- Complete CRUD operations (GET, POST, PUT, DELETE)
- Pagination support
- Request validation
- Error handling

**Exposed Endpoints:**
- `GET /api/v1/users`
- `GET /api/v1/users/:id`
- `POST /api/v1/users`
- `PUT /api/v1/users/:id`
- `DELETE /api/v1/users/:id`

---

## Dependencies

Added new dependency:
```bash
go get github.com/go-playground/validator/v10
```

---

## Usage

### 1. Success Response
```go
func GetUser(c echo.Context) error {
    user := getUserFromDB()
    return response.Success(c, user, "User retrieved")
}
```

**Output:**
```json
{
  "success": true,
  "message": "User retrieved",
  "data": { "id": "123", "name": "John" },
  "timestamp": 1672531200
}
```

### 2. Pagination Response
```go
func GetUsers(c echo.Context) error {
    var pagination response.PaginationRequest
    c.Bind(&pagination)
    
    users := fetchUsers(pagination.GetOffset(), pagination.GetPerPage())
    meta := response.CalculateMeta(pagination.GetPage(), pagination.GetPerPage(), 100)
    
    return response.SuccessWithMeta(c, users, meta)
}
```

Query: `GET /users?page=2&per_page=20`

**Output:**
```json
{
  "success": true,
  "data": [...],
  "meta": {
    "page": 2,
    "per_page": 20,
    "total": 100,
    "total_pages": 5
  },
  "timestamp": 1672531200
}
```

### 3. Request Validation
```go
type CreateUserRequest struct {
    Username string `json:"username" validate:"required,username"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"required,gte=18"`
}

func CreateUser(c echo.Context) error {
    var req CreateUserRequest
    
    if err := request.Bind(c, &req); err != nil {
        if validationErr, ok := err.(*request.ValidationError); ok {
            return response.ValidationError(c, "Validation failed", 
                validationErr.GetFieldErrors())
        }
        return response.BadRequest(c, err.Error())
    }
    
    user := createUser(req)
    return response.Created(c, user, "User created")
}
```

**Error Output (if validation fails):**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": {
      "username": "Username must be alphanumeric and 3-20 characters",
      "age": "age must be greater than or equal to 18"
    }
  },
  "timestamp": 1672531200
}
```

### 4. Error Responses
```go
// Not Found
return response.NotFound(c, "User not found")

// Unauthorized
return response.Unauthorized(c, "Invalid credentials")

// Bad Request
return response.BadRequest(c, "Invalid input")

// Internal Server Error
return response.InternalServerError(c, "Database error")
```

---

## Key Features

### 1. **Consistent Response Format**
All responses follow the same structure with fields `success`, `data`, `error`, `meta`, and `timestamp`.

### 2. **Built-in Validation**
Support for various validation rules:
- `required`, `email`, `min`, `max`, `len`
- `gte`, `lte`, `oneof`
- Custom: `phone`, `username`

### 3. **Pagination Helper**
```go
pagination.GetPage()      // Default: 1
pagination.GetPerPage()   // Default: 10, Max: 100
pagination.GetOffset()    // Calculate offset for DB query
```

### 4. **Comprehensive Error Handling**
Helper functions for all common HTTP status codes with customizable error details.

### 5. **Type Safe**
Uses Go structs for request and response, avoiding manual `map[string]interface{}`.

---

## Example Service Implementation

See [`service_a.go`](../internal/services/modules/service_a.go) for complete implementation example with:
- List with pagination
- Get single resource
- Create with validation
- Update with validation
- Delete with proper response

---

## Complete Documentation

See [`API_RESPONSE_STRUCTURE.md`](API_RESPONSE_STRUCTURE.md) for:
- Complete documentation of all functions
- Best practices
- Advanced examples
- Complete use cases

---

## Next Steps

1. **Use response helpers** in all service modules
2. **Implement validation** for all request structs
3. **Standardize error messages** across services
4. **Add custom validators** according to business needs
5. **Update existing endpoints** to use the new structure

---

## Build Status

```bash
- Dependencies installed
- go mod tidy completed
- Build successful
- Ready to use!
```

---

**Note:** This structure is production-ready and can be used immediately for all Echo services. All files are provided with complete documentation and examples.
