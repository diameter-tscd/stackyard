package services_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"stackyrd/internal/services/modules"
	"stackyrd/pkg/response"
	testhelpers "stackyrd/pkg/testing"
)

func TestNewUsersService(t *testing.T) {
	service := modules.NewUsersService(true)
	if service == nil {
		t.Fatal("expected service to be created")
	}
	if !service.Enabled() {
		t.Error("expected service to be enabled")
	}
	if service.Name() != "Users Service" {
		t.Errorf("expected name 'Users Service', got %q", service.Name())
	}
}

func TestUsersServiceDisabled(t *testing.T) {
	service := modules.NewUsersService(false)
	if service.Enabled() {
		t.Error("expected service to be disabled")
	}
}

func TestGetUsers(t *testing.T) {
	service := modules.NewUsersService(true)
	c, rec := testhelpers.NewTestContext(http.MethodGet, "/api/v1/users", nil)

	err := service.GetUsers(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testhelpers.AssertStatus(t, rec, http.StatusOK)

	var resp response.Response
	testhelpers.ParseResponse(t, rec, &resp)

	if !resp.Success {
		t.Error("expected success to be true")
	}
	if resp.Data == nil {
		t.Error("expected data to be present")
	}
}

func TestGetUsersWithPagination(t *testing.T) {
	service := modules.NewUsersService(true)
	c, rec := testhelpers.NewTestContextWithQuery(http.MethodGet, "/api/v1/users", map[string]string{
		"page":     "1",
		"per_page": "10",
	})

	err := service.GetUsers(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testhelpers.AssertStatus(t, rec, http.StatusOK)

	var resp response.Response
	testhelpers.ParseResponse(t, rec, &resp)

	if resp.Meta == nil {
		t.Error("expected meta to be present")
	}
}

func TestGetUser(t *testing.T) {
	service := modules.NewUsersService(true)
	c, rec := testhelpers.NewTestContextWithParams(http.MethodGet, "/api/v1/users/:id", map[string]string{
		"id": "1",
	}, nil)

	err := service.GetUser(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testhelpers.AssertStatus(t, rec, http.StatusOK)

	var resp response.Response
	testhelpers.ParseResponse(t, rec, &resp)

	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestGetUserNotFound(t *testing.T) {
	service := modules.NewUsersService(true)
	c, rec := testhelpers.NewTestContextWithParams(http.MethodGet, "/api/v1/users/:id", map[string]string{
		"id": "999",
	}, nil)

	err := service.GetUser(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testhelpers.AssertStatus(t, rec, http.StatusNotFound)

	var resp response.Response
	testhelpers.ParseResponse(t, rec, &resp)

	if resp.Success {
		t.Error("expected success to be false")
	}
}

func TestCreateUser(t *testing.T) {
	service := modules.NewUsersService(true)
	body := modules.CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		FullName: "Test User",
	}
	c, rec := testhelpers.NewTestContext(http.MethodPost, "/api/v1/users", body)

	err := service.CreateUser(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testhelpers.AssertStatus(t, rec, http.StatusCreated)

	var resp response.Response
	testhelpers.ParseResponse(t, rec, &resp)

	if !resp.Success {
		t.Error("expected success to be true")
	}
	if resp.Message != "User created successfully" {
		t.Errorf("expected message 'User created successfully', got %q", resp.Message)
	}
}

func TestCreateUserValidation(t *testing.T) {
	service := modules.NewUsersService(true)
	body := modules.CreateUserRequest{
		Username: "",
		Email:    "invalid-email",
		FullName: "T",
	}
	c, rec := testhelpers.NewTestContext(http.MethodPost, "/api/v1/users", body)

	err := service.CreateUser(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testhelpers.AssertStatus(t, rec, http.StatusUnprocessableEntity)

	var resp response.Response
	testhelpers.ParseResponse(t, rec, &resp)

	if resp.Success {
		t.Error("expected success to be false")
	}
}

func TestUpdateUser(t *testing.T) {
	service := modules.NewUsersService(true)
	body := modules.UpdateUserRequest{
		Username: "updateduser",
		Email:    "updated@example.com",
		FullName: "Updated User",
		Status:   "active",
	}
	c, rec := testhelpers.NewTestContextWithParams(http.MethodPut, "/api/v1/users/:id", map[string]string{
		"id": "1",
	}, body)

	err := service.UpdateUser(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testhelpers.AssertStatus(t, rec, http.StatusOK)

	var resp response.Response
	testhelpers.ParseResponse(t, rec, &resp)

	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestDeleteUser(t *testing.T) {
	service := modules.NewUsersService(true)
	c, rec := testhelpers.NewTestContextWithParams(http.MethodDelete, "/api/v1/users/:id", map[string]string{
		"id": "1",
	}, nil)

	err := service.DeleteUser(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testhelpers.AssertStatus(t, rec, http.StatusNoContent)
}

func TestDeleteUserNotFound(t *testing.T) {
	service := modules.NewUsersService(true)
	c, rec := testhelpers.NewTestContextWithParams(http.MethodDelete, "/api/v1/users/:id", map[string]string{
		"id": "999",
	}, nil)

	err := service.DeleteUser(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testhelpers.AssertStatus(t, rec, http.StatusNotFound)
}

func TestUserStruct(t *testing.T) {
	user := modules.User{
		ID:        "1",
		Username:  "testuser",
		Email:     "test@example.com",
		Status:    "active",
		CreatedAt: 1234567890,
	}

	jsonData, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("failed to marshal user: %v", err)
	}

	var decoded modules.User
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("failed to unmarshal user: %v", err)
	}

	if decoded.ID != user.ID {
		t.Errorf("expected ID %q, got %q", user.ID, decoded.ID)
	}
	if decoded.Username != user.Username {
		t.Errorf("expected Username %q, got %q", user.Username, decoded.Username)
	}
}

func BenchmarkGetUsers(b *testing.B) {
	service := modules.NewUsersService(true)
	for i := 0; i < b.N; i++ {
		c, _ := testhelpers.NewTestContext(http.MethodGet, "/api/v1/users", nil)
		_ = service.GetUsers(c)
	}
}

func BenchmarkGetUser(b *testing.B) {
	service := modules.NewUsersService(true)
	for i := 0; i < b.N; i++ {
		c, _ := testhelpers.NewTestContextWithParams(http.MethodGet, "/api/v1/users/:id", map[string]string{
			"id": "1",
		}, nil)
		_ = service.GetUser(c)
	}
}

func BenchmarkCreateUser(b *testing.B) {
	service := modules.NewUsersService(true)
	body := modules.CreateUserRequest{
		Username: "benchuser",
		Email:    "bench@example.com",
		FullName: "Benchmark User",
	}
	for i := 0; i < b.N; i++ {
		c, _ := testhelpers.NewTestContext(http.MethodPost, "/api/v1/users", body)
		_ = service.CreateUser(c)
	}
}
