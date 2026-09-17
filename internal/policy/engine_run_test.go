// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package policy_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/myshra777-ai/garuda/internal/policy"
)

// ─────────────────────────────────────────────────────────────────────────────
// Fixtures
// ─────────────────────────────────────────────────────────────────────────────

// openTestPool returns a pool connected to the test database. The
// test is skipped when neither TEST_DATABASE_URL nor DATABASE_URL is
// set — a unit test must never require a live database to pass.
func openTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL or DATABASE_URL must be set to run policy engine tests")
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Fatalf("ping: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// seedTenantWorkspace inserts a fresh tenant and workspace and
// registers cleanup. It assumes the minimal column set the rest of the
// codebase uses: tenants(id, name, created_at, updated_at) and
// workspaces(id, tenant_id, name, created_at, updated_at).
//
// The cleanup runs in LIFO order relative to openTestPool, so data is
// removed before the pool is closed.
func seedTenantWorkspace(t *testing.T, pool *pgxpool.Pool) (uuid.UUID, uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	tenantID := uuid.New()
	workspaceID := uuid.New()
	suffix := tenantID.String()[:8]

	if _, err := pool.Exec(ctx, `
		INSERT INTO tenants (id, name, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
	`, tenantID, "policy-test-"+suffix); err != nil {
		t.Fatalf("seed tenant: %v", err)
	}

	if _, err := pool.Exec(ctx, `
		INSERT INTO workspaces (id, tenant_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
	`, workspaceID, tenantID, "policy-test-ws-"+suffix); err != nil {
		_, _ = pool.Exec(ctx, `DELETE FROM tenants WHERE id = $1`, tenantID)
		t.Fatalf("seed workspace: %v", err)
	}

	t.Cleanup(func() {
		ctx := context.Background()
		_, _ = pool.Exec(ctx, `DELETE FROM policy_evaluations WHERE tenant_id = $1`, tenantID)
		_, _ = pool.Exec(ctx, `DELETE FROM policies WHERE tenant_id = $1`, tenantID)
		_, _ = pool.Exec(ctx, `DELETE FROM workspaces WHERE id = $1`, workspaceID)
		_, _ = pool.Exec(ctx, `DELETE FROM tenants WHERE id = $1`, tenantID)
	})

	return tenantID, workspaceID
}

// writeTestPolicies writes n ALLOW-only policies to a temp dir and
// returns the directory path. Each policy carries a single
// claim_exists predicate that cannot match in a fresh test tenant
// (there are no claims), so the engine evaluates it, finds no match,
// and produces an ALLOW evaluation with reason "no predicate
// matched". That is the minimum shape the parser accepts: a non-empty
// `when` list and a `then.reason` string.
func writeTestPolicies(t *testing.T, n int) string {
	t.Helper()
	dir := t.TempDir()
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("test-allow-%02d", i)
		body := fmt.Sprintf(`id: %s
version: v1
title: "Test Allow %02d"
description: Auto-generated test policy.
priority: 100
language: any
authority: "test@example.com"

scope:
  domain: test

when:
  - type: claim_exists
    params:
      claim_type: IMPORTS
      from_name_pattern: "^never-matches-.*$"
      to_name_pattern: "^never-matches-.*$"

then:
  decision: ALLOW
  reason: |
    Auto-generated test policy. Never matches in an empty test
    workspace; exists so the engine has a valid policy to evaluate.
`, id, i)
		path := filepath.Join(dir, id+".yaml")
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("write policy %d: %v", i, err)
		}
	}
	return dir
}

// evaluationCount returns the number of policy_evaluations rows for a
// tenant. This is the write that DryRun must suppress.
func evaluationCount(t *testing.T, pool *pgxpool.Pool, tenantID uuid.UUID) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(context.Background(),
		`SELECT COUNT(*)::int FROM policy_evaluations WHERE tenant_id = $1`,
		tenantID).Scan(&n); err != nil {
		t.Fatalf("count evaluations: %v", err)
	}
	return n
}

// listMerkleTables returns every table in the current schema whose
// name starts with "merkle". The engine's anchor writes to one of
// these; a dry run must not change any of them.
func listMerkleTables(t *testing.T, pool *pgxpool.Pool) []string {
	t.Helper()
	rows, err := pool.Query(context.Background(), `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = current_schema()
		  AND table_name LIKE 'merkle%'
		ORDER BY table_name
	`)
	if err != nil {
		t.Fatalf("discover merkle tables: %v", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan table name: %v", err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate merkle tables: %v", err)
	}
	return names
}

// merkleSnapshot returns the row count of every table in the current
// schema whose name starts with "merkle". Row counts, not max block
// heights, because the append-only ledger has no updatable state — an
// insert is the only mutation.
func merkleSnapshot(t *testing.T, pool *pgxpool.Pool) map[string]int {
	t.Helper()
	ctx := context.Background()

	// Materialize the table list first so the outer row set is closed
	// before the per-table COUNT queries open new ones.
	names := listMerkleTables(t, pool)

	out := make(map[string]int, len(names))
	for _, name := range names {
		quoted := pgx.Identifier{name}.Sanitize()
		var n int
		if err := pool.QueryRow(ctx, "SELECT COUNT(*)::int FROM "+quoted).Scan(&n); err != nil {
			t.Fatalf("count %s: %v", name, err)
		}
		out[name] = n
	}
	return out
}

// ─────────────────────────────────────────────────────────────────────────────
// Tests
// ─────────────────────────────────────────────────────────────────────────────

// TestRun_DryRunWritesNothing asserts that a dry-run pass leaves the
// database exactly as it found it: no new policy_evaluations rows, no
// new Merkle ledger entries.
//
// The dry run must still evaluate the policies so callers can preview
// decisions — the assertion is on the write path, not the read path.
func TestRun_DryRunWritesNothing(t *testing.T) {
	pool := openTestPool(t)
	tenantID, workspaceID := seedTenantWorkspace(t, pool)
	dir := writeTestPolicies(t, 3)

	// Baseline. Captured before the engine is even constructed so any
	// incidental write during setup would surface as a delta.
	evalBefore := evaluationCount(t, pool, tenantID)
	merkleBefore := merkleSnapshot(t, pool)

	engine := policy.NewEngine(pool)
	result, err := engine.Run(
		context.Background(),
		tenantID, workspaceID,
		dir,
		"manual", "test-subject", "test-actor",
		policy.RunOptions{DryRun: true, Reconcile: false},
	)
	if err != nil {
		t.Fatalf("Run (dry): %v", err)
	}

	// The engine must still return evaluations in a dry run.
	if !result.Preview {
		t.Errorf("RunResult.Preview = false, want true")
	}
	if got, want := len(result.Evaluations), 3; got != want {
		t.Errorf("len(Evaluations) = %d, want %d", got, want)
	}

	// Nothing written: policy_evaluations unchanged.
	if got := evaluationCount(t, pool, tenantID); got != evalBefore {
		t.Errorf("policy_evaluations count changed during dry run: before=%d after=%d", evalBefore, got)
	}

	// Nothing written: no Merkle table grew.
	merkleAfter := merkleSnapshot(t, pool)
	if !reflect.DeepEqual(merkleBefore, merkleAfter) {
		t.Errorf("Merkle ledger changed during dry run:\n  before=%v\n  after =%v",
			merkleBefore, merkleAfter)
	}
}

// TestRun_SaveWrites asserts that a non-dry-run pass writes exactly
// one policy_evaluations row per evaluated policy and anchors each to
// the Merkle ledger.
func TestRun_SaveWrites(t *testing.T) {
	pool := openTestPool(t)
	tenantID, workspaceID := seedTenantWorkspace(t, pool)
	dir := writeTestPolicies(t, 3)

	evalBefore := evaluationCount(t, pool, tenantID)

	engine := policy.NewEngine(pool)
	result, err := engine.Run(
		context.Background(),
		tenantID, workspaceID,
		dir,
		"manual", "test-subject", "test-actor",
		policy.RunOptions{DryRun: false, Reconcile: false},
	)
	if err != nil {
		t.Fatalf("Run (save): %v", err)
	}

	if result.Preview {
		t.Errorf("RunResult.Preview = true, want false")
	}

	// The engine evaluates every parsed policy that is not expired and
	// does not fail the language pre-filter. With language: any and no
	// expiry, that is all three.
	if got, want := len(result.Evaluations), 3; got != want {
		t.Errorf("len(Evaluations) = %d, want %d", got, want)
	}

	// Count must increase by exactly the number of evaluations the
	// engine reported writing. The engine skips a policy that fails
	// evaluation or persistence, and reports those skips by omitting
	// the evaluation from the result — so len(Evaluations) is the
	// authoritative delta, not len(parsed).
	evalAfter := evaluationCount(t, pool, tenantID)
	if got, want := evalAfter-evalBefore, len(result.Evaluations); got != want {
		t.Errorf("policy_evaluations delta = %d, want %d (before=%d after=%d)",
			got, want, evalBefore, evalAfter)
	}

	// Every persisted evaluation must carry a Merkle anchor. The
	// engine logs and continues on an anchor failure, so a NULL
	// merkle_block_height means the write path is not doing what its
	// contract claims.
	var unanchored int
	if err := pool.QueryRow(context.Background(), `
		SELECT COUNT(*)::int FROM policy_evaluations
		WHERE tenant_id = $1 AND merkle_block_height IS NULL
	`, tenantID).Scan(&unanchored); err != nil {
		t.Fatalf("count unanchored: %v", err)
	}
	if unanchored != 0 {
		t.Errorf("%d evaluations persisted without a Merkle anchor", unanchored)
	}
}
