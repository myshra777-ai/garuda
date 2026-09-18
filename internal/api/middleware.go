// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/auth"
)

// responseWriter wraps http.ResponseWriter to capture status code accurately.
type responseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.statusCode = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK) // Capture implicit success codes
	}
	return rw.ResponseWriter.Write(b)
}

// WithRequestID adds a unique request ID to the context and response headers.

// WithLogging logs each request with method, path, status, and duration.
func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(ww, r)
		duration := time.Since(start)
		slog.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.statusCode,
			"duration_ms", duration.Milliseconds(),
			"request_id", r.Context().Value("request_id"),
		)
	})
}

// WithRecovery catches panics in handlers and returns 500.
func WithRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered", "error", err, "request_id", r.Context().Value("request_id"))
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// WithRateLimit applies rate limiting per IP address.
func WithRateLimit(limiter *IPRateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if !limiter.Allow(ip) {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// WithCORS adds CORS headers to responses.
func WithCORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Garuda-Actor, X-Request-ID, Authorization")
					break
				}
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// WithMerkleHeader injects the latest active Merkle root hash into outgoing HTTP response headers.
func (s *Server) WithMerkleHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID, err := resolveTenantID(r, uuid.Nil)
		if err == nil && tenantID != uuid.Nil {
			if snap, err := s.store.GetLatestMerkleSnapshot(r.Context(), tenantID); err == nil && snap != nil {
				w.Header().Set("X-Garuda-Merkle-Root", snap.SnapshotHash)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// WithAuth wraps an http.Handler with JWT authentication middleware using the provided JWTConfig.
// WithAuth wraps an http.Handler with JWT authentication middleware using the provided JWTConfig.
// WithAuth wraps an http.Handler with JWT authentication middleware using the provided JWTConfig.
func WithAuth(jwtConfig *auth.JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := ""
			authHeader := r.Header.Get("Authorization")
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenStr = authHeader[7:]
			}
			if tokenStr == "" {
				tokenStr = r.URL.Query().Get("token")
			}

			if tokenStr == "" {
				http.Error(w, `{"error":"unauthorized: missing authorization token"}`, http.StatusUnauthorized)
				return
			}

			actor, tenantID, err := jwtConfig.ValidateToken(tokenStr)
			if err != nil {
				http.Error(w, `{"error":"unauthorized: invalid token"}`, http.StatusUnauthorized)
				return
			}

			ctx := auth.ContextWithActorAndTenant(r.Context(), actor, fmt.Sprintf("%v", tenantID))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// WithRequestID adds a request ID to the context and response headers.
func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		// Set response header
		w.Header().Set("X-Request-ID", requestID)
		// Add to context
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
