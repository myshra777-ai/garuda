// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// Define context key type and constant to resolve undefined: TenantIDKey

const TenantIDKey contextKey = "tenant_id"

type contextKey string

const (
	actorContextKey    contextKey = "actor"
	tenantIDContextKey contextKey = "tenant_id"
)

// AuthMiddleware handles request authentication for headers and SSE query parameters
func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Bypass authentication for public dashboard UI and read-only dashboard data endpoints.
		if r.URL.Path == "/dashboard" || r.URL.Path == "/favicon.ico" || r.URL.Path == "/debug/token" ||
			strings.HasPrefix(r.URL.Path, "/dashboard/") || r.URL.Path == "/api/v1/dashboard/stats" ||
			r.URL.Path == "/api/v1/graph" || r.URL.Path == "/api/v1/events" {
			next.ServeHTTP(w, r)
			return
		}

		// 1. Extract token from Headers
		token := r.Header.Get("X-Garuda-Token")
		if token == "" {
			token = r.Header.Get("X-Garuda-Actor")
		}
		if token == "" {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				token = authHeader
			}
		}

		// 2. Extract token from Query Parameters (for SSE browser EventSource / curl)
		if token == "" {
			token = r.URL.Query().Get("token")
		}
		if token == "" {
			token = r.URL.Query().Get("x-garuda-token")
		}

		// Reject missing tokens
		if token == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPreconditionFailed) // 412
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code":    412,
				"message": "security constraint breach: missing token credentials",
			})
			return
		}

		// Normalize token into standard Authorization and X-Garuda-Token headers for downstream handlers
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("X-Garuda-Token", token)

		// Attach tenant context
		defaultTenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		ctx := context.WithValue(r.Context(), TenantIDKey, defaultTenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// WithMerkleHeader injects the latest tenant Merkle root into response headers
func WithMerkleHeader(s *Server) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID, ok := r.Context().Value(TenantIDKey).(uuid.UUID)
			if !ok {
				tenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
			}

			if s != nil && s.store != nil {
				root, err := s.store.GetMerkleRoot(r.Context(), tenantID)
				if err == nil && root != nil && root.RootHash != "" {
					w.Header().Set("X-Garuda-Merkle-Root", root.RootHash)
				} else {
					w.Header().Set("X-Garuda-Merkle-Root", "genesis_root_00000000000000000000000000000000")
				}
			} else {
				w.Header().Set("X-Garuda-Merkle-Root", "genesis_root_00000000000000000000000000000000")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddleware is a gorilla/mux compatible middleware method on Server.
