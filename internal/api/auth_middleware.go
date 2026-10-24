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
	"github.com/myshra777-ai/garuda/internal/auth"
	"github.com/myshra777-ai/garuda/internal/tenant"
)

// Define context key type and constant to resolve undefined: TenantIDKey

const TenantIDKey contextKey = "tenant_id"

type contextKey string

const (
	actorContextKey    contextKey = "actor"
	tenantIDContextKey contextKey = "tenant_id"
)

// AuthMiddleware handles request authentication for headers, session cookies, and SSE query parameters.
// When s.jwtConfig is configured, Bearer/X-Garuda-Token tokens are cryptographically verified as
// Ed25519 JWTs via s.jwtConfig.ValidateToken, and tenant_id + actor are derived from the verified claims.
func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Bypass authentication for public dashboard UI and read-only dashboard data endpoints.
		if r.URL.Path == "/dashboard" || r.URL.Path == "/favicon.ico" || r.URL.Path == "/debug/token" ||
			strings.HasPrefix(r.URL.Path, "/dashboard/") || r.URL.Path == "/api/v1/dashboard/stats" ||
			r.URL.Path == "/api/v1/graph" || r.URL.Path == "/api/v1/events" {
			next.ServeHTTP(w, r)
			return
		}

		// 0. Check valid browser session cookie first (for Web UI / Dashboard API calls)
		if s != nil && s.sessions != nil {
			if c, err := r.Cookie(SessionCookieName); err == nil && c.Value != "" {
				if sess, ok := s.sessions.Get(c.Value); ok {
					defaultTenantID := tenant.CanonicalID
					ctx := context.WithValue(r.Context(), TenantIDKey, defaultTenantID)
					ctx = auth.ContextWithActorAndTenant(ctx, sess.UserEmail, defaultTenantID)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		// 1. Extract token from Headers
		token := strings.TrimSpace(r.Header.Get("X-Garuda-Token"))
		if token == "" {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			} else if strings.HasPrefix(authHeader, "bearer ") {
				token = strings.TrimSpace(strings.TrimPrefix(authHeader, "bearer "))
			} else if authHeader != "" {
				token = authHeader
			}
		}
		if token == "" {
			token = strings.TrimSpace(r.Header.Get("X-Garuda-Actor"))
		}

		// 2. Extract token from Query Parameters (for SSE browser EventSource / curl)
		if token == "" {
			token = strings.TrimSpace(r.URL.Query().Get("token"))
		}
		if token == "" {
			token = strings.TrimSpace(r.URL.Query().Get("x-garuda-token"))
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

		// 3. Cryptographically verify Ed25519 JWT when s.jwtConfig is configured
		resolvedTenantID := tenant.CanonicalID
		resolvedActor := strings.TrimSpace(r.Header.Get("X-Garuda-Actor"))
		if s != nil && s.jwtConfig != nil {
			actor, tokenTenantID, err := s.jwtConfig.ValidateToken(token)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized) // 401
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code":    401,
					"message": "unauthorized: invalid or expired Ed25519 JWT token",
				})
				return
			}
			resolvedActor = actor
			if tokenTenantID != uuid.Nil {
				resolvedTenantID = tokenTenantID
			}
		}

		// Normalize token into standard Authorization and X-Garuda-Token headers for downstream handlers
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("X-Garuda-Token", token)
		if resolvedActor != "" {
			r.Header.Set("X-Garuda-Actor", resolvedActor)
		}

		// Attach verified tenant and actor context
		ctx := context.WithValue(r.Context(), TenantIDKey, resolvedTenantID)
		ctx = auth.ContextWithActorAndTenant(ctx, resolvedActor, resolvedTenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// WithMerkleHeader injects the latest tenant Merkle root into response headers
func WithMerkleHeader(s *Server) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID, ok := r.Context().Value(TenantIDKey).(uuid.UUID)
			if !ok {
				tenantID = tenant.CanonicalID
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
