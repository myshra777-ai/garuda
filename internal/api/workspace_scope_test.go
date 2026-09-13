// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// TestResolveWorkspaceForUser_Isolation covers the Session C access
// rule: a user may only resolve a workspace they are a member of.
//
// Setup creates two independent tenants via the signup flow. Each
// tenant has a default workspace named "default". The test then
// verifies:
//
//  1. A user asking for a workspace name that exists only in another
//     tenant receives ErrNoWorkspaceAccess. Cross-tenant case.
//
//  2. A user asking for a workspace in their own tenant receives a
//     scope with the correct IDs and role.
//
//  3. An empty name resolves to the caller's own workspace, never to
//     a workspace from another tenant.
func TestResolveWorkspaceForUser_Isolation(t *testing.T) {
	pgStore := getTestPostgresStore(t)
	ctx := context.Background()

	suffix := uuid.New().String()[:8]
	aliceEmail := "alice-iso-" + suffix + "@local"
	bobEmail := "bob-iso-" + suffix + "@local"
	hash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = pgStore.Pool().Exec(cleanupCtx,
			`DELETE FROM users WHERE email = ANY($1)`,
			[]string{aliceEmail, bobEmail})
		_, _ = pgStore.Pool().Exec(cleanupCtx,
			`DELETE FROM tenants WHERE name LIKE $1`,
			"%"+suffix+"%")
	})

	alice, aliceWS, err := pgStore.SignupUser(ctx, aliceEmail, hash, "Alice Iso")
	if err != nil {
		t.Fatalf("signup alice: %v", err)
	}
	bob, _, err := pgStore.SignupUser(ctx, bobEmail, hash, "Bob Iso")
	if err != nil {
		t.Fatalf("signup bob: %v", err)
	}

	aliceID, _ := uuid.Parse(alice.ID)
	bobID, _ := uuid.Parse(bob.ID)

	// Rename Alice's workspace to something unique to her tenant.
	const aliceOnlyName = "alice-only-workspace"
	if _, err := pgStore.Pool().Exec(ctx,
		`UPDATE workspaces SET name = $1 WHERE id = $2`,
		aliceOnlyName, aliceWS.ID); err != nil {
		t.Fatalf("rename alice workspace: %v", err)
	}

	// Case 1: Bob asks for Alice's workspace. Must fail closed with
	// ErrNoWorkspaceAccess. A 200 here would be a cross-tenant leak.
	if _, err := resolveWorkspaceForUser(ctx, pgStore, bobID, bobEmail, "user", aliceOnlyName); !errors.Is(err, ErrNoWorkspaceAccess) {
		t.Fatalf("bob asking for alice's workspace: err = %v; want ErrNoWorkspaceAccess", err)
	}

	// Case 2: Alice asks for her own workspace. Must succeed with the
	// correct IDs and role.
	scope, err := resolveWorkspaceForUser(ctx, pgStore, aliceID, aliceEmail, "user", aliceOnlyName)
	if err != nil {
		t.Fatalf("alice asking for her own workspace: %v", err)
	}
	if scope.WorkspaceID != aliceWS.ID {
		t.Fatalf("scope.WorkspaceID = %s; want %s", scope.WorkspaceID, aliceWS.ID)
	}
	if scope.TenantID != aliceWS.TenantID {
		t.Fatalf("scope.TenantID = %s; want %s", scope.TenantID, aliceWS.TenantID)
	}
	if scope.WorkspaceRole != "owner" {
		t.Fatalf("scope.WorkspaceRole = %q; want %q", scope.WorkspaceRole, "owner")
	}

	// Case 3: Alice's empty-name lookup must land in her own tenant,
	// never in another.
	aliceFallback, err := resolveWorkspaceForUser(ctx, pgStore, aliceID, aliceEmail, "user", "")
	if err != nil {
		t.Fatalf("alice empty-name: %v", err)
	}
	if aliceFallback.TenantID != aliceWS.TenantID {
		t.Fatalf("alice empty-name scope has tenant %s; want %s (her own tenant)",
			aliceFallback.TenantID, aliceWS.TenantID)
	}

	// Case 4: Bob's empty-name lookup must land in Bob's tenant, not Alice's.
	bobFallback, err := resolveWorkspaceForUser(ctx, pgStore, bobID, bobEmail, "user", "")
	if err != nil {
		t.Fatalf("bob empty-name: %v", err)
	}
	if bobFallback.TenantID == aliceWS.TenantID {
		t.Fatalf("bob empty-name landed in alice's tenant %s", aliceWS.TenantID)
	}
}
