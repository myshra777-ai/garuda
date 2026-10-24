// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/myshra777-ai/garuda/internal/auth"
)

func TestRouterRegistersBrowserRoutes(t *testing.T) {
	s := &Server{
		rateLimiter: NewIPRateLimiter(100, 100),
		sessions:    NewSessionStore(0),
	}
	router := mux.NewRouter()
	s.RegisterRoutes(router)

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/login"},
		{http.MethodPost, "/login"},
		{http.MethodGet, "/signup"},
		{http.MethodPost, "/signup"},
		{http.MethodPost, "/logout"},
		{http.MethodGet, "/dashboard"},
		{http.MethodGet, "/admin"},
		{http.MethodGet, "/api/v1/workspaces"},
		{http.MethodGet, "/api/v1/dashboard/stats"},
		{http.MethodGet, "/api/v1/dashboard/search"},
		{http.MethodGet, "/api/v1/dashboard/policies"},
		{http.MethodGet, "/api/v1/graph"},
		{http.MethodGet, "/health"},
		{http.MethodGet, "/healthz"},
		{http.MethodGet, "/readyz"},
		{http.MethodGet, "/api/v1/system/health"},
		{http.MethodPost, "/api/v1/telemetry/spans"},
	}

	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			match := &mux.RouteMatch{}
			if !router.Match(req, match) {
				t.Errorf("%s %s not registered on the router", tc.method, tc.path)
			}
		})
	}
}

func TestAuthMiddleware_EnforcesRealEd25519JWTValidation(t *testing.T) {
	jwtCfg, err := auth.NewJWTConfig("garuda-test", "garuda-api", 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to create JWTConfig: %v", err)
	}

	s := &Server{
		jwtConfig:   jwtCfg,
		rateLimiter: NewIPRateLimiter(100, 100),
		sessions:    NewSessionStore(0),
	}

	var capturedActor string
	var capturedTenant uuid.UUID

	protected := s.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedActor, _ = auth.ActorFromContext(r.Context())
		capturedTenant, _ = auth.TenantFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	// 1. Forged / arbitrary non-empty Bearer string MUST be rejected with 401
	badReq := httptest.NewRequest(http.MethodPost, "/api/v1/decisions/propose", nil)
	badReq.Header.Set("Authorization", "Bearer arbitrary-forged-token")
	badRec := httptest.NewRecorder()
	protected.ServeHTTP(badRec, badReq)
	if badRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for forged token, got %d", badRec.Code)
	}

	// 2. Validly signed Ed25519 JWT MUST be accepted and populate actor + tenant claims
	expectedTenant := uuid.New()
	validToken, err := jwtCfg.GenerateToken("secops@acme.io", expectedTenant)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	goodReq := httptest.NewRequest(http.MethodPost, "/api/v1/decisions/propose", nil)
	goodReq.Header.Set("Authorization", "Bearer "+validToken)
	goodRec := httptest.NewRecorder()
	protected.ServeHTTP(goodRec, goodReq)
	if goodRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for valid Ed25519 JWT, got %d: %s", goodRec.Code, goodRec.Body.String())
	}
	if capturedActor != "secops@acme.io" {
		t.Fatalf("expected actor secops@acme.io from JWT claims, got %q", capturedActor)
	}
	if capturedTenant != expectedTenant {
		t.Fatalf("expected tenant %s from JWT claims, got %s", expectedTenant, capturedTenant)
	}
}
