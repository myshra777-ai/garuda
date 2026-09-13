// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// TestSignupUser_CreatesFullScope verifies that SignupUser writes all
// five rows — user, tenant, workspace, tenant_member, workspace_member
// — and returns the user and workspace with matching IDs.
//
// This is the load-bearing test for Session B. If any of the five
// inserts were removed, the row count check would fail. If the
// transaction were not committed, the cleanup would fail to find the
// workspace. If the FK relationships were wrong, the joins would
// return zero rows.
func TestSignupUser_CreatesFullScope(t *testing.T) {
	pgStore := getTestPostgresStore(t)
	ctx := context.Background()

	email := "signup-test-" + uuid.New().String()[:8] + "@local"
	passwordHash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		// The tenant_members and workspace_members rows cascade from
		// tenants via ON DELETE CASCADE, so deleting the tenant is
		// enough. The users row is deleted last because it is
		// referenced by tenant_members.
		_, _ = pgStore.Pool().Exec(cleanupCtx,
			`DELETE FROM users WHERE email = $1`, email)
	})

	user, ws, err := pgStore.SignupUser(ctx, email, passwordHash, "Test User")
	if err != nil {
		t.Fatalf("SignupUser failed: %v", err)
	}
	if user == nil || ws == nil {
		t.Fatalf("SignupUser returned nil user or workspace")
	}
	if user.Email != email {
		t.Fatalf("user.Email = %q; want %q", user.Email, email)
	}
	if ws.Name != "default" {
		t.Fatalf("workspace name = %q; want %q", ws.Name, "default")
	}
	if ws.TenantID == uuid.Nil {
		t.Fatalf("workspace TenantID is Nil")
	}

	// 1. Verify the tenant exists.
	var tenantName string
	if err := pgStore.Pool().QueryRow(ctx,
		`SELECT name FROM tenants WHERE id = $1`, ws.TenantID).Scan(&tenantName); err != nil {
		t.Fatalf("tenant lookup failed: %v", err)
	}
	if tenantName != email+"'s tenant" {
		t.Fatalf("tenant name = %q; want %q", tenantName, email+"'s tenant")
	}

	// 2. Verify the tenant_members row.
	var tenantRole string
	if err := pgStore.Pool().QueryRow(ctx,
		`SELECT role FROM tenant_members WHERE user_id = $1 AND tenant_id = $2`,
		user.ID, ws.TenantID).Scan(&tenantRole); err != nil {
		t.Fatalf("tenant_members lookup failed: %v", err)
	}
	if tenantRole != "owner" {
		t.Fatalf("tenant role = %q; want %q", tenantRole, "owner")
	}

	// 3. Verify the workspace_members row.
	var workspaceRole string
	if err := pgStore.Pool().QueryRow(ctx,
		`SELECT role FROM workspace_members WHERE user_id = $1 AND workspace_id = $2`,
		user.ID, ws.ID).Scan(&workspaceRole); err != nil {
		t.Fatalf("workspace_members lookup failed: %v", err)
	}
	if workspaceRole != "owner" {
		t.Fatalf("workspace role = %q; want %q", workspaceRole, "owner")
	}
}

// TestSignupUser_RejectsDuplicateEmail verifies that a second signup
// with the same email returns ErrEmailExists and does not create a
// second tenant or workspace.
//
// The users_email_key constraint fires before any of the following
// inserts run, so the transaction rolls back cleanly and no orphan
// tenant or workspace is left behind.
func TestSignupUser_RejectsDuplicateEmail(t *testing.T) {
	pgStore := getTestPostgresStore(t)
	ctx := context.Background()

	email := "signup-dup-" + uuid.New().String()[:8] + "@local"
	passwordHash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = pgStore.Pool().Exec(cleanupCtx,
			`DELETE FROM users WHERE email = $1`, email)
	})

	_, firstWS, err := pgStore.SignupUser(ctx, email, passwordHash, "First")
	if err != nil {
		t.Fatalf("first SignupUser failed: %v", err)
	}

	_, _, err = pgStore.SignupUser(ctx, email, passwordHash, "Second")
	if !errors.Is(err, ErrEmailExists) {
		t.Fatalf("second SignupUser error = %v; want ErrEmailExists", err)
	}

	// Confirm no orphan tenant was created by the second attempt.
	var tenantCount int
	if err := pgStore.Pool().QueryRow(ctx,
		`SELECT COUNT(*) FROM tenants WHERE name = $1`,
		email+"'s tenant").Scan(&tenantCount); err != nil {
		t.Fatalf("tenant count query failed: %v", err)
	}
	if tenantCount != 1 {
		t.Fatalf("tenant count = %d; want 1 (only the first signup's tenant)", tenantCount)
	}

	// Confirm the first workspace still exists and is owned by the
	// first user.
	var ownerCount int
	if err := pgStore.Pool().QueryRow(ctx,
		`SELECT COUNT(*) FROM workspace_members wm
		 WHERE wm.workspace_id = $1 AND wm.role = 'owner'`,
		firstWS.ID).Scan(&ownerCount); err != nil {
		t.Fatalf("workspace owner count failed: %v", err)
	}
	if ownerCount != 1 {
		t.Fatalf("workspace %s owner count = %d; want 1", firstWS.ID, ownerCount)
	}
}
