// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

// querySampleCount is the ring capacity. 10,000 samples is roughly the
// last few minutes of activity on a busy server, or a few hours on an
// idle one. Either is enough for a stable percentile.
const querySampleCount = 10000

// minSamplesForPercentile is the threshold below which the read path
// reports "collecting samples" rather than a fabricated P50. A
// percentile over fewer than 10 queries is noise.
const minSamplesForPercentile = 10

// queryRing is an in-process ring buffer of query durations. Reset on
// process restart, which is correct — uptime is also per-process, and
// the two numbers are read together.
type queryRing struct {
	mu      sync.Mutex
	samples []time.Duration
	idx     int
	count   int
}

var globalQueryRing = &queryRing{
	samples: make([]time.Duration, querySampleCount),
}

// QueryTracer implements pgx.QueryTracer. Attached to the store pool's
// ConnConfig, it records every query's duration into the ring.
type QueryTracer struct{}

type queryStartKey struct{}

func (QueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	return context.WithValue(ctx, queryStartKey{}, time.Now())
}

func (QueryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryEndData) {
	start, ok := ctx.Value(queryStartKey{}).(time.Time)
	if !ok {
		return
	}
	globalQueryRing.record(time.Since(start))
}

func (r *queryRing) record(d time.Duration) {
	r.mu.Lock()
	r.samples[r.idx] = d
	r.idx = (r.idx + 1) % querySampleCount
	if r.count < querySampleCount {
		r.count++
	}
	r.mu.Unlock()
}

// QueryLatencyPercentiles returns P50, P95, and P99 in milliseconds,
// plus the number of samples currently in the ring.
//
// When fewer than minSamplesForPercentile samples exist, the three
// values are zero and the caller renders "Collecting samples" rather
// than a number. Returns (0, 0, 0, 0) when no samples exist at all.
//
// The ring is copied under the lock and sorted outside it. Sorting
// 10,000 items takes about 10ms; the Operations tab loads once per
// refresh, so the cost is invisible.
func QueryLatencyPercentiles() (p50, p95, p99 float64, samples int) {
	r := globalQueryRing
	r.mu.Lock()
	n := r.count
	if n == 0 {
		r.mu.Unlock()
		return 0, 0, 0, 0
	}
	snapshot := make([]time.Duration, n)
	copy(snapshot, r.samples[:n])
	r.mu.Unlock()

	if n < minSamplesForPercentile {
		return 0, 0, 0, n
	}

	sort.Slice(snapshot, func(i, j int) bool { return snapshot[i] < snapshot[j] })

	pct := func(p float64) float64 {
		idx := int(float64(n-1) * p)
		return float64(snapshot[idx].Microseconds()) / 1000.0
	}

	return pct(0.50), pct(0.95), pct(0.99), n
}
