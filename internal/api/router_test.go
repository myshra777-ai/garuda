// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/myshra777-ai/garuda/internal/api"
	"github.com/myshra777-ai/garuda/internal/auth"
)

func TestSetupRouter_OrchestrationRoutesRegistered(t *testing.T) {
	srv := api.NewServer(nil, nil, nil, nil, nil, nil, nil)
	router := api.SetupRouter(srv, srv.RateLimiter(), nil, nil)

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/agent/checkpoint"},
		{http.MethodPost, "/api/v1/agent/handoff"},
		{http.MethodGet, "/api/v1/agent/handoff/test-session-id"},
		{http.MethodPost, "/api/v1/agent/resume"},
		{http.MethodGet, "/api/v1/lineage/status"},
		{http.MethodGet, "/api/v1/system/health"},
		{http.MethodGet, "/healthz"},
		{http.MethodGet, "/readyz"},
	}

	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			var match mux.RouteMatch
			if !router.Match(req, &match) {
				t.Fatalf("route %s %s is not registered on the router", tc.method, tc.path)
			}
			if match.MatchErr == mux.ErrMethodMismatch {
				t.Fatalf("route %s %s matched path but rejected method", tc.method, tc.path)
			}
		})
	}
}

func TestSetupRouter_OverrideReplacesDefault(t *testing.T) {
	srv := api.NewServer(nil, nil, nil, nil, nil, nil, nil)

	overrideCalled := false
	router := api.SetupRouter(srv, srv.RateLimiter(), nil, func(r *mux.Router) {
		r.HandleFunc("/api/v1/graph", func(w http.ResponseWriter, r *http.Request) {
			overrideCalled = true
			w.WriteHeader(http.StatusTeapot)
		}).Methods(http.MethodGet)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/graph", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if !overrideCalled {
		t.Fatalf("expected override handler to be called for /api/v1/graph, got status %d", rec.Code)
	}
	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected status %d from override, got %d", http.StatusTeapot, rec.Code)
	}
}

func TestAuthMiddleware_EnforcesRealEd25519JWTValidation(t *testing.T) {
	jwtCfg, err := auth.NewJWTConfig("garuda-test", "garuda-api", 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to create JWTConfig: %v", err)
	}

	srv := api.NewServer(nil, nil, jwtCfg, nil, nil, nil, nil)
	var capturedActor string
	var capturedTenant uuid.UUID

	protected := srv.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedActor, _ = auth.ActorFromContext(r.Context())
		capturedTenant, _ = auth.TenantFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	// 1. Arbitrary non-empty Bearer string MUST be rejected with 401
	badReq := httptest.NewRequest(http.MethodGet, "/api/v1/decisions", nil)
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

	goodReq := httptest.NewRequest(http.MethodGet, "/api/v1/decisions", nil)
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
