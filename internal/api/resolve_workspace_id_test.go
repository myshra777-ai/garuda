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

// TestResolveWorkspaceID_TenantScoped proves that resolveWorkspaceID
// cannot return a workspace belonging to a tenant other than the
// dashboard tenant.
//
// Before the fix, the named lookup ran `SELECT ... WHERE name = $1
// LIMIT 1` with no tenant filter. Since UNIQUE (tenant_id, name) allows
// the same name under multiple tenants, and since the sixteen
// 'workspace-core' rows that existed before migration 073 were exactly
// that pattern, a caller could receive a workspace from any tenant.
// The empty-name branch had the same class of bug: ORDER BY updated_at
// DESC over an unscoped workspaces table.
//
// The test inserts the same workspace name under two tenants — the
// canonical one and a fresh one — and asserts that both lookups return
// a row owned by the canonical tenant.
func TestResolveWorkspaceID_TenantScoped(t *testing.T) {
	pgStore := getTestPostgresStore(t)
	ctx := context.Background()

	canonicalTenant := dashboardTenantUUID
	otherTenant := uuid.New()
	sharedName := "resolve-ws-test-" + uuid.New().String()[:8]

	t.Cleanup(func() {
		// Deletes both tenants' rows. t.Cleanup fires on t.Fatal and on
		// panic, so residue never survives a failed run.
		cleanupCtx := context.Background()
		_, _ = pgStore.Pool().Exec(cleanupCtx,
			`DELETE FROM workspaces WHERE name = $1`, sharedName)
	})

	// Insert the non-canonical row first so that a query without a
	// tenant filter has a chance of returning it. Insertion order does
	// not matter for correctness, but it removes any ambiguity about
	// which row a broken query would have found "first" if the planner
	// happened to prefer insertion order.
	if _, err := pgStore.CreateWorkspace(ctx, otherTenant.String(), sharedName, ""); err != nil {
		t.Fatalf("CreateWorkspace (other tenant) failed: %v", err)
	}
	canonicalWS, err := pgStore.CreateWorkspace(ctx, canonicalTenant.String(), sharedName, "")
	if err != nil {
		t.Fatalf("CreateWorkspace (canonical tenant) failed: %v", err)
	}

	// --- Named lookup must return the canonical row. ---
	gotID, gotName, err := resolveWorkspaceID(ctx, pgStore, sharedName)
	if err != nil {
		t.Fatalf("resolveWorkspaceID(%q) failed: %v", sharedName, err)
	}
	if gotID != canonicalWS.ID {
		t.Fatalf("resolveWorkspaceID(%q) returned workspace %s; want the canonical-tenant workspace %s. "+
			"This means the lookup is not filtering by tenant_id and can leak across tenants.",
			sharedName, gotID, canonicalWS.ID)
	}
	if gotName != sharedName {
		t.Fatalf("resolveWorkspaceID(%q) returned name %q; want %q", sharedName, gotName, sharedName)
	}

	// --- Empty-name lookup must return a canonical-tenant workspace. ---
	//
	// The test does not assert which workspace is returned — the
	// canonical tenant may have other rows with a more recent
	// updated_at. It asserts only that whatever is returned belongs to
	// the canonical tenant, which is the property the fix establishes.
	fallbackID, _, err := resolveWorkspaceID(ctx, pgStore, "")
	if err != nil {
		t.Fatalf("resolveWorkspaceID(\"\") failed: %v", err)
	}
	var tenantOfReturned uuid.UUID
	if err := pgStore.Pool().QueryRow(ctx,
		`SELECT tenant_id FROM workspaces WHERE id = $1`, fallbackID).Scan(&tenantOfReturned); err != nil {
		t.Fatalf("could not look up tenant of returned workspace: %v", err)
	}
	if tenantOfReturned != canonicalTenant {
		t.Fatalf("resolveWorkspaceID(\"\") returned workspace owned by tenant %s; want canonical %s",
			tenantOfReturned, canonicalTenant)
	}
}
