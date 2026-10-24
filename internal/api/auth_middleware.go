// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/auth"
	"github.com/myshra777-ai/garuda/internal/tenant"
)

type contextKey string

const TenantIDKey contextKey = "tenant_id"

// AuthMiddleware enforces cryptographic authentication across all /api/v1 routes.
//
// Authentication precedence:
//  1. Active browser session cookie (garuda_session) verified against s.sessions.
//  2. Cryptographically signed Ed25519 JWT passed via Authorization: Bearer <jwt>,
//     X-Garuda-Token: <jwt>, or ?token=<jwt>, verified via s.jwtConfig.ValidateToken.
//
// Tenant ID and Actor identity are derived directly from the verified JWT claims
// (or authenticated browser session), never from unverified client headers when
// JWTConfig is configured.
func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defaultTenantID := tenant.CanonicalID

		// 1. Check valid browser session cookie first (for Web UI / Dashboard / SSE calls)
		if s.sessions != nil {
			if c, err := r.Cookie(SessionCookieName); err == nil && c.Value != "" {
				if sess, ok := s.sessions.Get(c.Value); ok {
					ctx := context.WithValue(r.Context(), TenantIDKey, defaultTenantID)
					ctx = auth.ContextWithActorAndTenant(ctx, sess.UserEmail, defaultTenantID)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		// 2. Extract Bearer token, X-Garuda-Token header, or ?token= query param
		var token string
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				token = strings.TrimSpace(parts[1])
			} else {
				writeJSONError(w, "invalid Authorization header format; expected 'Bearer <token>'", http.StatusUnauthorized)
				return
			}
		}
		if token == "" {
			token = strings.TrimSpace(r.Header.Get("X-Garuda-Token"))
		}
		if token == "" {
			token = strings.TrimSpace(r.URL.Query().Get("token"))
		}

		if token == "" {
			writeJSONError(w, "unauthorized: valid Ed25519 Bearer JWT or session cookie required", http.StatusUnauthorized)
			return
		}

		// 3. When JWTConfig is present (all production and dev daemon instances),
		// cryptographically validate the Ed25519 signature, expiration, and claims.
		if s.jwtConfig != nil {
			actor, tokenTenantID, err := s.jwtConfig.ValidateToken(token)
			if err != nil {
				writeJSONError(w, "unauthorized: invalid or expired JWT token", http.StatusUnauthorized)
				return
			}
			if tokenTenantID == uuid.Nil {
				tokenTenantID = defaultTenantID
			}
			ctx := context.WithValue(r.Context(), TenantIDKey, tokenTenantID)
			ctx = auth.ContextWithActorAndTenant(ctx, actor, tokenTenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// 4. Fail closed if running in production without JWTConfig initialized
		if os.Getenv("GARUDA_ENV") == "production" {
			writeJSONError(w, "internal error: JWT verifier not initialized", http.StatusInternalServerError)
			return
		}

		// 5. Unit-test harness fallback (only when Server is constructed with nil jwtConfig in tests)
		actor := strings.TrimSpace(r.Header.Get("X-Garuda-Actor"))
		if actor == "" {
			actor = "cli-agent"
		}
		ctx := context.WithValue(r.Context(), TenantIDKey, defaultTenantID)
		ctx = auth.ContextWithActorAndTenant(ctx, actor, defaultTenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
