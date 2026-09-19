// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"net/http"
	"sync"
	"time"
)

// rateCounter tracks total and error responses over a rolling
// 60-minute window, bucketed per minute. In-process only. Cleared on
// process restart, which is correct: uptime is also per-process.
//
// "Error" means a 5xx response. 4xx is client-side and not counted
// — the operator cares about server failures, not user mistakes.
//
// Concurrency: Record is called once per request from the middleware.
// Snapshot is called on Operations-tab page loads. Both take the
// mutex; the critical section is small (a few integer writes).
type rateCounter struct {
	mu            sync.Mutex
	buckets       [60]rateBucket
	currentMinute int64 // unix minute; 0 until first Record
	currentIdx    int
}

type rateBucket struct {
	minute int64
	total  int64
	errors int64
}

// Record increments the current minute's counters. Must be called
// exactly once per completed request.
func (r *rateCounter) Record(status int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().Unix() / 60
	r.advanceLocked(now)

	b := &r.buckets[r.currentIdx]
	b.total++
	if status >= 500 {
		b.errors++
	}
}

// Snapshot returns the totals over the last 60 minutes.
func (r *rateCounter) Snapshot() (total, errors int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().Unix() / 60
	r.advanceLocked(now)

	for i := range r.buckets {
		total += r.buckets[i].total
		errors += r.buckets[i].errors
	}
	return
}

// advanceLocked rotates the bucket ring forward to cover the current
// minute. If the gap is 60 minutes or more, every bucket is stale, so
// the whole ring is reset instead of walked one minute at a time.
func (r *rateCounter) advanceLocked(now int64) {
	if r.currentMinute == 0 {
		r.currentMinute = now
		r.currentIdx = 0
		r.buckets[0] = rateBucket{minute: now}
		return
	}
	if now <= r.currentMinute {
		return
	}
	gap := now - r.currentMinute
	if gap >= 60 {
		r.buckets = [60]rateBucket{}
		r.currentIdx = 0
		r.buckets[0] = rateBucket{minute: now}
		r.currentMinute = now
		return
	}
	for i := int64(0); i < gap; i++ {
		r.currentMinute++
		r.currentIdx = (r.currentIdx + 1) % 60
		r.buckets[r.currentIdx] = rateBucket{minute: r.currentMinute}
	}
}

// statusRecorder wraps http.ResponseWriter to capture the status code
// without changing the response. It preserves http.Flusher, which the
// SSE endpoints require.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (s *statusRecorder) WriteHeader(code int) {
	if s.wroteHeader {
		return
	}
	s.status = code
	s.wroteHeader = true
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if !s.wroteHeader {
		s.status = http.StatusOK
		s.wroteHeader = true
	}
	return s.ResponseWriter.Write(b)
}

// Flush forwards to the underlying writer if it supports flushing.
// Required for SSE endpoints; without it, streaming responses buffer
// indefinitely.
func (s *statusRecorder) Flush() {
	if f, ok := s.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// WithErrorRateTracking records every request's status into the
// counter. Placed after WithRequestID so every counted request has a
// request ID, and outside the rate limiter so 429s are counted as
// traffic (not as errors).
func WithErrorRateTracking(counter *rateCounter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			counter.Record(rec.status)
		})
	}
}
