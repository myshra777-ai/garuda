// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// SetupRouter composes the complete HTTP handler chain.
//
// One router, one middleware chain, one route table. Every entry
// point that serves HTTP — the production API server, the dev
// daemon, integration tests — calls this function. A route added to
// RegisterRoutes is served everywhere without a second edit.
func SetupRouter(
	server *Server,
	rateLimiter *IPRateLimiter,
	bridgeHandler http.HandlerFunc,
	extraRoutes func(*mux.Router),
) http.Handler {
	router := mux.NewRouter()

	// Canonical route table. See handler.go:RegisterRoutes.
	server.RegisterRoutes(router)

	// Entry-point-specific routes. Registered after RegisterRoutes
	// so they take precedence (gorilla matches the most recently
	// registered route for a path). The dev daemon uses this for
	// its graph visualizer override. Production passes nil.
	if extraRoutes != nil {
		extraRoutes(router)
	}

	// Production-only routes not part of the canonical table.
	router.HandleFunc("/docs", server.HandleSwaggerUI).Methods(http.MethodGet)
	router.HandleFunc("/openapi.yaml", server.HandleOpenAPISpec).Methods(http.MethodGet)
	router.HandleFunc("/openapi.json", server.HandleOpenAPISpec).Methods(http.MethodGet)
	if bridgeHandler != nil {
		router.HandleFunc("/mcp/bridge", bridgeHandler).Methods(http.MethodPost)
	}

	router.PathPrefix("/static/").Handler(server.HandleStatic())

	// Middleware chain, outermost first. Rate limit runs before any
	// route, so unauthenticated floods are dropped before they reach
	// auth code. RegisterRoutes no longer applies rate limiting; it
	// lives here, once.
	return WithRecovery(
		WithLogging(
			WithRequestID(
				server.WithMerkleHeader(
					WithRateLimit(rateLimiter)(
						WithCORS([]string{"*"})(router),
					),
				),
			),
		),
	)
}
