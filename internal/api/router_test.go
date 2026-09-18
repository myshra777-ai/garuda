// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

// TestRouterRegistersBrowserRoutes asserts that the routes a browser
// client depends on are registered on the canonical route table.
//
// This test exists because a prior refactor moved the browser routes
// (/login, /signup, /logout, /api/v1/workspaces, /admin) onto a
// gorilla subrouter that was never mounted on the served handler.
// Every one of those routes returned 404 in the deployed binary.
// The /dashboard page loaded but its workspace picker silently
// failed, and no external tester could complete signup.
//
// The test does not dispatch requests — that requires a fully
// constructed Server with an initialized session store and JWT
// config. It asserts that the routes are registered on the router,
// which is the property the previous refactor violated.
func TestRouterRegistersBrowserRoutes(t *testing.T) {
	s := &Server{
		// RegisterRoutes calls s.RateLimitMiddleware through
		// r.Use, which dereferences s.rateLimiter. Provide a real one.
		rateLimiter: NewIPRateLimiter(100, 100),
		// sessions must be non-nil for the session middleware
		// constructor, but the middleware is not invoked during
		// registration, so an empty store is sufficient.
		sessions: NewSessionStore(0),
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
