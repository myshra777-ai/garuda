// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package refund

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter.
type RateLimiter struct {
	mu       sync.Mutex
	tokens   map[string]*tokenBucket
	capacity int
	refill   time.Duration
}

type tokenBucket struct {
	tokens     int
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter with the given capacity and refill duration.
func NewRateLimiter(capacity int, refill time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:   make(map[string]*tokenBucket),
		capacity: capacity,
		refill:   refill,
	}
}

// Allow checks if a request from the given IP is allowed.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.tokens[ip]
	if !exists {
		bucket = &tokenBucket{tokens: rl.capacity, lastRefill: time.Now()}
		rl.tokens[ip] = bucket
	}

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
