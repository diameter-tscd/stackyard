package services_test

import (
	"testing"

	"stackyard/internal/services/modules"
	testhelpers "stackyard/pkg/testing"
)

func TestNewProductsService(t *testing.T) {
	service := modules.NewProductsService(true)
	if service == nil {
		t.Fatal("expected service to be created")
	}
	if !service.Enabled() {
		t.Error("expected service to be enabled")
	}
	if service.Name() != "Products Service" {
		t.Errorf("expected name 'Products Service', got %q", service.Name())
	}
	if service.WireName() != "products-service" {
		t.Errorf("expected wire name 'products-service', got %q", service.WireName())
	}
}

func TestProductsServiceDisabled(t *testing.T) {
	service := modules.NewProductsService(false)
	if service.Enabled() {
		t.Error("expected service to be disabled")
	}
}

func TestProductsServiceEndpoints(t *testing.T) {
	service := modules.NewProductsService(true)
	endpoints := service.Endpoints()
	if len(endpoints) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(endpoints))
	}
	if endpoints[0] != "/products" {
		t.Errorf("expected endpoint '/products', got %q", endpoints[0])
	}
}

func TestProductsServiceRegisterRoutes(t *testing.T) {
	service := modules.NewProductsService(true)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RegisterRoutes panicked: %v", r)
		}
	}()

	e := testhelpers.NewTestEcho()
	g := e.Group("/api/v1")
	service.RegisterRoutes(g)
}

func BenchmarkProductsServiceName(b *testing.B) {
	service := modules.NewProductsService(true)
	for i := 0; i < b.N; i++ {
		_ = service.Name()
	}
}

func BenchmarkProductsServiceEnabled(b *testing.B) {
	service := modules.NewProductsService(true)
	for i := 0; i < b.N; i++ {
		_ = service.Enabled()
	}
}
