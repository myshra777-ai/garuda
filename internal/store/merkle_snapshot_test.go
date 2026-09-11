// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// TestWorkerSnapshot_EmptyTenantSucceeds proves that a tenant with no
// entities and no claim verifications can produce a valid v0 workspace
// snapshot that satisfies migration 069's format CHECK constraints.
//
// Before this test, the empty-static-tier path wrote the sentinel
// string "GARUDA_EMPTY_STATIC_TREE" into static_root_hash, which fails
// the 64-char hex CHECK.
func TestWorkerSnapshot_EmptyTenantSucceeds(t *testing.T) {
	s, cleanup := testDB(t)
	defer cleanup()

	tenantID := uuid.New()
	ctx := context.Background()

	// Ensure no rows for this tenant.
	_, _ = s.pool.Exec(ctx, `DELETE FROM merkle_snapshots WHERE tenant_id = $1`, tenantID)
	_, _ = s.pool.Exec(ctx, `DELETE FROM merkle_roots WHERE tenant_id = $1`, tenantID)
	_, _ = s.pool.Exec(ctx, `DELETE FROM merkle_epoch_leaves WHERE tenant_id = $1`, tenantID)
	defer func() {
		_, _ = s.pool.Exec(ctx, `DELETE FROM merkle_snapshots WHERE tenant_id = $1`, tenantID)
		_, _ = s.pool.Exec(ctx, `DELETE FROM merkle_roots WHERE tenant_id = $1`, tenantID)
		_, _ = s.pool.Exec(ctx, `DELETE FROM merkle_epoch_leaves WHERE tenant_id = $1`, tenantID)
	}()

	snap, err := s.CreateUnifiedMerkleSnapshot(ctx, tenantID)
	if err != nil {
		t.Fatalf("empty tenant snapshot: %v", err)
	}
	if snap.BlockHeight != 1 {
		t.Errorf("block height = %d, want 1", snap.BlockHeight)
	}
	if snap.SnapshotHash == "" {
		t.Error("snapshot hash is empty")
	}

	// Read back and confirm verification_version is 0.
	var version int
	err = s.pool.QueryRow(ctx, `
		SELECT verification_version FROM merkle_snapshots WHERE tenant_id = $1
	`, tenantID).Scan(&version)
	if err != nil {
		t.Fatalf("read version: %v", err)
	}
	if version != 0 {
		t.Errorf("verification_version = %d, want 0 for worker snapshot", version)
	}

	// Confirm the empty-tree hash was used, not a sentinel string.
	var staticRoot string
	err = s.pool.QueryRow(ctx, `
		SELECT static_root_hash FROM merkle_snapshots WHERE tenant_id = $1
	`, tenantID).Scan(&staticRoot)
	if err != nil {
		t.Fatalf("read static root: %v", err)
	}
	if len(staticRoot) != 64 {
		t.Errorf("static_root_hash length = %d, want 64 hex chars", len(staticRoot))
	}
}
