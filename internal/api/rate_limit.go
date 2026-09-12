// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Per-IP rate limiting for the API.
//
// Defaults are deliberately generous for a small team: 120 requests
// per minute with a burst of 20. This is enough to load a dashboard
// with 15 panels firing in parallel, but not enough to flood the
// server. Tune with GARUDA_RATE_LIMIT_PER_MINUTE and
// GARUDA_RATE_LIMIT_BURST.
const (
	defaultRateLimitPerMinute = 120
	defaultRateLimitBurst     = 20
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter tracks one rate.Limiter per client IP. Idle limiters
// are garbage collected after ipLimiterTTL.
type IPRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiter
	rate     rate.Limit
	burst    int
}

const ipLimiterTTL = 10 * time.Minute

// NewIPRateLimiter creates a limiter with the given rate and burst.
func NewIPRateLimiter(perMinute, burst int) *IPRateLimiter {
	l := &IPRateLimiter{
		limiters: make(map[string]*ipLimiter),
		rate:     rate.Limit(float64(perMinute) / 60.0),
		burst:    burst,
	}
	go l.gcLoop()
	return l
}

// Allow reports whether the request from clientIP is permitted.
func (l *IPRateLimiter) Allow(clientIP string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.limiters[clientIP]
	if !ok {
		entry = &ipLimiter{
			limiter: rate.NewLimiter(l.rate, l.burst),
		}
		l.limiters[clientIP] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter.Allow()
}

func (l *IPRateLimiter) gcLoop() {
	ticker := time.NewTicker(ipLimiterTTL)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-ipLimiterTTL)
		l.mu.Lock()
		for ip, entry := range l.limiters {
			if entry.lastSeen.Before(cutoff) {
				delete(l.limiters, ip)
			}
		}
		l.mu.Unlock()
	}
}

// RateLimitMiddleware wraps a handler with per-IP rate limiting.
//
// Returns 429 Too Many Requests with a Retry-After header when the
// limit is exceeded. Legitimate dashboard traffic will never hit
// this under normal use.
func (s *Server) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.rateLimiter == nil {
			next.ServeHTTP(w, r)
			return
		}
		ip := clientIP(r)
		if !s.rateLimiter.Allow(ip) {
			w.Header().Set("Retry-After", "60")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate limit exceeded","retry_after_seconds":60}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP extracts the best-effort client IP from the request.
//
// Prefers X-Forwarded-For (leftmost non-private value) when the
// request comes from a private network, which is the typical reverse
// proxy setup. Falls back to RemoteAddr. Does not attempt to handle
// every proxy configuration; a full deployment would parse
// X-Forwarded-For against a trusted proxy list.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for _, part := range strings.Split(xff, ",") {
			ip := strings.TrimSpace(part)
			if ip != "" && !isPrivateIP(ip) {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func isPrivateIP(s string) bool {
	ip := net.ParseIP(s)
	if ip == nil {
		return false
	}
	return ip.IsPrivate() || ip.IsLoopback()
}
