// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"log/slog"
	"time"
)

func (s *PostgresStore) StartExpiryWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.markExpiredTasks(ctx); err != nil {
				slog.Error("failed to mark expired tasks", "error", err)
			}
		}
	}
}

func (s *PostgresStore) markExpiredTasks(ctx context.Context) error {
	query := `
		UPDATE decisions
		SET status = 'stale', updated_at = NOW()
		WHERE temporal_metadata->>'expires_at' < NOW()::TEXT
		AND status != 'executed' AND status != 'stale';
	`
	_, err := s.pool.Exec(ctx, query)
	return err
}
