// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
)

// TestCreateWorkspace_ConcurrentSameName proves that N goroutines calling
// CreateWorkspace with the same (tenant_id, name) all return without
// error, all return the same workspace ID, and exactly one of them sees
// inserted=true.
//
// Before the fix, the second caller raced the first and returned
// "failed to execute create workspace query: no rows in result set":
// the CTE's ON CONFLICT DO NOTHING suppressed the RETURNING row, and
// the UNION ALL fallback SELECT ran under the same READ COMMITTED
// statement snapshot, which did not include the first caller's
// uncommitted insert.
func TestCreateWorkspace_ConcurrentSameName(t *testing.T) {
	pgStore := getTestPostgresStore(t)
	ctx := context.Background()

	tenantID := uuid.New()
	name := "concurrent-create-test-" + uuid.New().String()[:8]

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = pgStore.Pool().Exec(cleanupCtx,
			`DELETE FROM workspaces WHERE tenant_id = $1`, tenantID)
	})

	const N = 10
	start := make(chan struct{})
	results := make([]struct {
		ws       *Workspace
		err      error
		inserted bool
	}, N)

	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func(idx int) {
			defer wg.Done()
			<-start // barrier: fire all at once

			ws, err := pgStore.CreateWorkspace(ctx, tenantID.String(), name, "")

			// inserted = (err == nil && ws != nil). The
			// ErrWorkspaceExists path returns the existing workspace
			// alongside the error, so we record both.
			inserted := err == nil
			results[idx].ws = ws
			results[idx].err = err
			results[idx].inserted = inserted
		}(i)
	}
	close(start)
	wg.Wait()

	// Every caller must succeed. Either err == nil (this caller
	// inserted) or err == ErrWorkspaceExists (another caller won).
	var winners []*Workspace
	insertedCount := 0
	for i, r := range results {
		if r.err != nil && !errors.Is(r.err, ErrWorkspaceExists) {
			t.Fatalf("goroutine %d: unexpected error: %v", i, r.err)
		}
		if r.ws == nil {
			t.Fatalf("goroutine %d: nil workspace on non-nil error", i)
		}
		if r.inserted {
			insertedCount++
		}
		winners = append(winners, r.ws)
	}

	if insertedCount != 1 {
		t.Fatalf("expected exactly 1 caller with inserted=true, got %d", insertedCount)
	}

	// All callers must have returned the same workspace ID.
	firstID := winners[0].ID
	for i, w := range winners {
		if w.ID != firstID {
			t.Fatalf("goroutine %d: workspace ID = %s; want %s (all callers must agree)",
				i, w.ID, firstID)
		}
	}
}
