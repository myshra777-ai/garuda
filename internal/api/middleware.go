package api

import (
	"container/list"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
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

// LRUCache implements a fixed-size LRU cache for the rate limiter.
type lruCache struct {
	mu       sync.Mutex
	capacity int
	items    map[string]*list.Element
	order    *list.List
}

type cacheItem struct {
	key   string
	value *tokenBucket
}

func newLRUCache(capacity int) *lruCache {
	return &lruCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

func (c *lruCache) get(key string) (*tokenBucket, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		return elem.Value.(*cacheItem).value, true
	}
	return nil, false
}

func (c *lruCache) set(key string, value *tokenBucket) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		elem.Value.(*cacheItem).value = value
		return
	}
	if len(c.items) >= c.capacity {
		// Evict least recently used
		oldest := c.order.Back()
		if oldest != nil {
			c.order.Remove(oldest)
			delete(c.items, oldest.Value.(*cacheItem).key)
		}
	}
	elem := c.order.PushFront(&cacheItem{key: key, value: value})
	c.items[key] = elem
}

// RateLimiter implements a token bucket with LRU eviction.
type RateLimiter struct {
	cache    *lruCache
	capacity int
	refill   time.Duration
}

type tokenBucket struct {
	tokens     int
	lastRefill time.Time
}

func NewRateLimiter(capacity int, refill time.Duration, maxIPs int) *RateLimiter {
	return &RateLimiter{
		cache:    newLRUCache(maxIPs),
		capacity: capacity,
		refill:   refill,
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	bucket, exists := rl.cache.get(ip)
	if !exists {
		bucket = &tokenBucket{tokens: rl.capacity, lastRefill: time.Now()}
		rl.cache.set(ip, bucket)
	}

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill)
	bucket.lastRefill = now
	bucket.tokens += int(elapsed / rl.refill)
	if bucket.tokens > rl.capacity {
		bucket.tokens = rl.capacity
	}

	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}
	return false
}

// WithRateLimit applies rate limiting per IP address.
func WithRateLimit(limiter *RateLimiter) func(http.Handler) http.Handler {
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
