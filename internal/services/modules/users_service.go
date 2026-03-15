package modules

import (
	"stackyard/config"
	"stackyard/pkg/interfaces"
	"stackyard/pkg/logger"
	"stackyard/pkg/registry"
	"stackyard/pkg/request"
	"stackyard/pkg/response"
	"time"

	"github.com/labstack/echo/v4"
)

type UsersService struct {
	enabled bool
}

func NewUsersService(enabled bool) *UsersService {
	return &UsersService{enabled: enabled}
}

func (s *UsersService) Name() string        { return "Users Service" }
func (s *UsersService) Enabled() bool       { return s.enabled }
func (s *UsersService) Endpoints() []string { return []string{"/users", "/users/:id"} }

func (s *UsersService) RegisterRoutes(g *echo.Group) {
	sub := g.Group("/users")

	// List users with pagination
	sub.GET("", s.GetUsers)

	// Get single user
	sub.GET("/:id", s.GetUser)

	// Create user
	sub.POST("", s.CreateUser)

	// Update user
	sub.PUT("/:id", s.UpdateUser)

	// Delete user
	sub.DELETE("/:id", s.DeleteUser)
}

// Sample User struct
type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"created_at"`
}

// Request structs
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,username"`
	Email    string `json:"email" validate:"required,email"`
	FullName string `json:"full_name" validate:"required,min=3,max=100"`
}

type UpdateUserRequest struct {
	Username string `json:"username" validate:"omitempty,username"`
	Email    string `json:"email" validate:"omitempty,email"`
	FullName string `json:"full_name" validate:"omitempty,min=3,max=100"`
	Status   string `json:"status" validate:"omitempty,oneof=active inactive suspended"`
}

// Handlers

func (s *UsersService) GetUsers(c echo.Context) error {
	// Parse pagination from query
	var pagination response.PaginationRequest
	if err := c.Bind(&pagination); err != nil {
		return response.BadRequest(c, "Invalid pagination parameters")
	}

	// Mock data
	users := []User{
		{ID: "1", Username: "john_doe", Email: "john@example.com", Status: "active", CreatedAt: time.Now().Unix()},
		{ID: "2", Username: "jane_smith", Email: "jane@example.com", Status: "active", CreatedAt: time.Now().Unix()},
		{ID: "3", Username: "bob_wilson", Email: "bob@example.com", Status: "inactive", CreatedAt: time.Now().Unix()},
	}

	// Calculate metadata
	total := int64(len(users))
	meta := response.CalculateMeta(
		pagination.GetPage(),
		pagination.GetPerPage(),
		total,
	)

	return response.SuccessWithMeta(c, users, meta, "Users retrieved successfully")
}

func (s *UsersService) GetUser(c echo.Context) error {
	id := c.Param("id")

	// Mock data - in real app, fetch from database
	user := User{
		ID:        id,
		Username:  "john_doe",
		Email:     "john@example.com",
		Status:    "active",
		CreatedAt: time.Now().Unix(),
	}

	// Simulate not found
	if id == "999" {
		return response.NotFound(c, "User not found")
	}

	return response.Success(c, user, "User retrieved successfully")
}

func (s *UsersService) CreateUser(c echo.Context) error {
	var req CreateUserRequest

	// Bind and validate
	if err := request.Bind(c, &req); err != nil {
		if validationErr, ok := err.(*request.ValidationError); ok {
			return response.ValidationError(c, "Validation failed", validationErr.GetFieldErrors())
		}
		return response.BadRequest(c, err.Error())
	}

	// Mock user creation
	user := User{
		ID:        "123",
		Username:  req.Username,
		Email:     req.Email,
		Status:    "active",
		CreatedAt: time.Now().Unix(),
	}

	return response.Created(c, user, "User created successfully")
}

func (s *UsersService) UpdateUser(c echo.Context) error {
	id := c.Param("id")

	var req UpdateUserRequest

	// Bind and validate
	if err := request.Bind(c, &req); err != nil {
		if validationErr, ok := err.(*request.ValidationError); ok {
			return response.ValidationError(c, "Validation failed", validationErr.GetFieldErrors())
		}
		return response.BadRequest(c, err.Error())
	}

	// Mock updated user
	user := User{
		ID:        id,
		Username:  req.Username,
		Email:     req.Email,
		Status:    req.Status,
		CreatedAt: time.Now().Unix(),
	}

	return response.Success(c, user, "User updated successfully")
}

func (s *UsersService) DeleteUser(c echo.Context) error {
	id := c.Param("id")

	// Simulate not found
	if id == "999" {
		return response.NotFound(c, "User not found")
	}

	// Mock deletion - in real app, delete from database
	// No content response
	return response.NoContent(c)
}

// Auto-registration function - called when package is imported
func init() {
	registry.RegisterService("users_service", func(config *config.Config, logger *logger.Logger, deps *registry.Dependencies) interfaces.Service {
		return NewUsersService(config.Services.IsEnabled("users_service"))
	})
}
