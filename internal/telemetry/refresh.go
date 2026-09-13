// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package telemetry

import (
	"context"
	"log/slog"
	"time"

	"github.com/myshra777-ai/garuda/internal/store"
)

// RefreshInterval is how often the aggregate refresh runs.
//
// Fifteen minutes is the freshness bar. A SaaS dashboard that shows
// yesterday's counts only shows numbers nobody can act on. Fifteen
// minutes is fast enough that a signup appears before the operator
// finishes making coffee, and slow enough that the aggregate query
// does not compete with request traffic.
const RefreshInterval = 15 * time.Minute

// StartAggregateRefresher runs RefreshAggregates on start for today
// and yesterday, then on a ticker for today.
//
// The startup pass covers the overnight gap. If the daemon was
// stopped at 23:50 and restarted at 09:00, no ticker fired during
// the night; without the startup pass, yesterday's aggregate would
// be missing until the ticker fires at 09:15 and picks up today only.
//
// Returns immediately. The loop runs in a goroutine and exits when
// ctx is cancelled.
func StartAggregateRefresher(ctx context.Context, store *store.PostgresStore) {
	now := time.Now().UTC()
	yesterday := now.Add(-24 * time.Hour)

	if err := store.RefreshAggregates(ctx, yesterday); err != nil {
		slog.Warn("aggregate refresh failed", "date", yesterday.Format("2006-01-02"), "error", err)
	}
	if err := store.RefreshAggregates(ctx, now); err != nil {
		slog.Warn("aggregate refresh failed", "date", now.Format("2006-01-02"), "error", err)
	}

	go func() {
		ticker := time.NewTicker(RefreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				today := time.Now().UTC()
				if err := store.RefreshAggregates(ctx, today); err != nil {
					slog.Warn("aggregate refresh failed", "date", today.Format("2006-01-02"), "error", err)
				}
			}
		}
	}()
}
