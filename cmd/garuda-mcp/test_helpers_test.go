// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/store"
)

// getTestMCPServer returns an MCPServer backed by a real database, or
// skips the test if no database is configured. Only the `store` field
// is set: the handlers in this test file touch store and
// resolveTenantAndWorkspace, and nothing else.
//
// Mirrors getTestPostgresStore in internal/store/test_helpers_test.go.
// The two helpers cannot be shared because they live in different
// packages and the store helper is unexported.
func getTestMCPServer(t *testing.T) *MCPServer {
	t.Helper()

	dbURL := os.Getenv("GARUDA_TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		t.Skip("skipping DB-backed MCP test: neither GARUDA_TEST_DATABASE_URL nor DATABASE_URL is set")
	}

	s, err := store.NewPostgresStore(dbURL)
	if err != nil {
		t.Fatalf("failed to initialize postgres store: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	return &MCPServer{store: s}
}

// seedTenantWorkspace inserts one tenant and one workspace. The MCP
// handlers call resolveTenantAndWorkspace, which requires the tenant
// to have at least one workspace to fall back to when no workspace
// name is given in args. The workspace is not used by the
// handoff/resume paths; it exists only so resolution succeeds.
func seedTenantWorkspace(t *testing.T, srv *MCPServer, tenantID uuid.UUID, workspaceName string) {
	t.Helper()
	ctx := context.Background()

	if _, err := srv.store.Pool().Exec(ctx, `
		INSERT INTO tenants (id, name) VALUES ($1, $2)
		ON CONFLICT (id) DO NOTHING
	`, tenantID, "mcp-test-"+tenantID.String()[:8]); err != nil {
		t.Fatalf("seed tenant: %v", err)
	}

	if _, err := srv.store.Pool().Exec(ctx, `
		INSERT INTO workspaces (id, tenant_id, name) VALUES ($1, $2, $3)
		ON CONFLICT (tenant_id, name) DO NOTHING
	`, uuid.New(), tenantID, workspaceName); err != nil {
		t.Fatalf("seed workspace: %v", err)
	}
}

// cleanupTenantFixtures removes every fixture for one tenant. Delete
// in reverse dependency order: handoffs references agent_checkpoints,
// agents, and tasks; tasks references agents; workspaces references
// tenants.
func cleanupTenantFixtures(t *testing.T, srv *MCPServer, tenantID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	_, _ = srv.store.Pool().Exec(ctx, `DELETE FROM handoffs WHERE tenant_id = $1`, tenantID)
	_, _ = srv.store.Pool().Exec(ctx, `DELETE FROM agent_checkpoints WHERE tenant_id = $1`, tenantID)
	_, _ = srv.store.Pool().Exec(ctx, `DELETE FROM tasks WHERE tenant_id = $1`, tenantID)
	_, _ = srv.store.Pool().Exec(ctx, `DELETE FROM agents WHERE tenant_id = $1`, tenantID)
	_, _ = srv.store.Pool().Exec(ctx, `DELETE FROM workspaces WHERE tenant_id = $1`, tenantID)
	_, _ = srv.store.Pool().Exec(ctx, `DELETE FROM tenants WHERE id = $1`, tenantID)
}
