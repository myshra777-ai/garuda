// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// TestComputeDrift_SurvivesWorkspaceRename proves that the drift panel
// does not return zero after a workspace is renamed.
//
// Before migration 075, computeDrift filtered document_claims by the
// workspace name (string). Renaming the workspace left the string in
// document_claims stale, so the join silently matched nothing and the
// drift panel rendered zero across every category. The metric read
// "no drift" while the truth was "the lookup key moved."
//
// The test inserts one document claim under a fresh workspace, asserts
// computeDrift returns a non-zero count, renames the workspace, and
// asserts computeDrift still returns the same non-zero count.
func TestComputeDrift_SurvivesWorkspaceRename(t *testing.T) {
	pgStore := getTestPostgresStore(t)
	ctx := context.Background()

	tenantID := dashboardTenantUUID
	wsName := "drift-rename-test-" + uuid.New().String()[:8]

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = pgStore.Pool().Exec(cleanupCtx,
			`DELETE FROM document_claims WHERE document_title = 'drift-rename-test'`)
		_, _ = pgStore.Pool().Exec(cleanupCtx,
			`DELETE FROM workspaces WHERE name LIKE 'drift-rename-test-%' OR name LIKE 'drift-rename-test-renamed-%'`)
	})

	ws, err := pgStore.CreateWorkspace(ctx, tenantID.String(), wsName, "")
	if err != nil {
		t.Fatalf("CreateWorkspace failed: %v", err)
	}

	// Insert one document claim scoped by workspace_id.
	_, err = pgStore.Pool().Exec(ctx, `
		INSERT INTO document_claims (
			workspace, workspace_id, tenant_id, document_path, document_title,
			document_type, line_start, line_end, raw_statement,
			subject, modality, predicate, object
		) VALUES (
			$1, $2, $3, '/fake/doc.md', 'drift-rename-test',
			'ADR', 1, 1, 'test claim',
			'test-subject', 'MUST', 'CALLS', 'test-object'
		)
	`, wsName, ws.ID, tenantID)
	if err != nil {
		t.Fatalf("insert document_claim failed: %v", err)
	}

	// Baseline: 1 claim visible under the workspace.
	d1 := computeDrift(ctx, pgStore, ws.ID)
	if d1.TotalDocumentClaims != 1 {
		t.Fatalf("before rename: TotalDocumentClaims = %d; want 1", d1.TotalDocumentClaims)
	}

	// Rename the workspace.
	newName := "drift-rename-test-renamed-" + uuid.New().String()[:8]
	_, err = pgStore.Pool().Exec(ctx,
		`UPDATE workspaces SET name = $1 WHERE id = $2`, newName, ws.ID)
	if err != nil {
		t.Fatalf("rename failed: %v", err)
	}

	// The drift panel must still see the claim. Before the fix this
	// returned 0 because the query filtered on the stale name.
	d2 := computeDrift(ctx, pgStore, ws.ID)
	if d2.TotalDocumentClaims != 1 {
		t.Fatalf("after rename: TotalDocumentClaims = %d; want 1. "+
			"computeDrift is still keyed on the workspace name.",
			d2.TotalDocumentClaims)
	}
}
