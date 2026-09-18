// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"testing"

	"github.com/myshra777-ai/garuda/internal/api"
)

func TestRouterRegistrationNoPanics(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SetupRouter panicked during route registration: %v", r)
		}
	}()

	rateLimiter := api.NewIPRateLimiter(100, 100)
	server := &api.Server{} // zero-value server, registration-only test

	handler := api.SetupRouter(server, rateLimiter, nil, nil)
	if handler == nil {
		t.Fatal("SetupRouter returned nil")
	}

	// Route behavior is tested in internal/api/router_test.go, where
	// the Server struct's unexported fields (jwtConfig, sessions) can
	// be constructed. From this package the server is zero-valued and
	// cannot satisfy the auth middleware.
}
