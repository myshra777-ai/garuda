package main

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/myshra777-ai/garuda/internal/api"
	"github.com/myshra777-ai/garuda/internal/auth"
)

func TestRouterRegistrationNoPanics(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SetupRouter panicked during route registration: %v", r)
		}
	}()

	jwtConfig, err := auth.NewJWTConfig("test", "test", 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to create JWT config: %v", err)
	}
	rateLimiter := api.NewRateLimiter(100, time.Minute, 1000)
	server := &api.Server{} // mock server

	handler := SetupRouter(server, jwtConfig, rateLimiter)
	if handler == nil {
		t.Fatal("SetupRouter returned nil")
	}

	// Test a few routes
	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/health"},
		{"GET", "/system/discover"},
		{"POST", "/api/v1/decisions/submit"},
		{"GET", "/api/v1/plan"},
	}

	for _, route := range routes {
		req := httptest.NewRequest(route.method, route.path, nil)
		rec := httptest.NewRecorder()
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Dispatching %s %s panicked: %v", route.method, route.path, r)
				}
			}()
			handler.ServeHTTP(rec, req)
		}()
	}
}
