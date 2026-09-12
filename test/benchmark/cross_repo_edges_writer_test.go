// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package benchmark

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/analyzer"
)

// TestCrossRepoEdges_WriterPersistsEdges proves that SaveSemanticGraph
// writes rows to cross_repo_edges when one repository imports a package
// belonging to another repository in the same workspace.
//
// This test exists because before commit 3687c90 the resolved branch of
// detectCrossRepoImports used a bare ON CONFLICT target that PostgreSQL
// cannot bind to a partial unique index. Every resolved cross-repo edge
// threw SQLSTATE 42P10, was caught by slog.Warn, and was swallowed.
// cross_repo_edges was empty for the entire life of the table. No test
// in the package exercised SaveSemanticGraph before this one, so the
// failure was never observed.
//
// Setup:
//   - fresh tenant and workspace
//   - two repositories with distinct module paths in that workspace
//   - repo B provides a package entity whose package_path, package, and
//     name all equal its module path, so findPackageEntityInRepo resolves
//     it as the target of an IMPORTS edge
//   - repo A provides a file entity whose file_path matches the
//     Evidence.File of a synthetic IMPORTS relationship targeting repo
//     B's module path
//   - both repositories are added before any SaveSemanticGraph call so
//     the module map cache sees them both on first population
//
// Assertion:
//   - cross_repo_edges has at least one row scoped to the fresh
//     workspace
//   - the row is on the resolved branch (resolved = true), which is the
//     branch that was broken
//
// Cleanup deletes symbol_cache, repositories, and workspaces for the
// fresh tenant in that order. t.Cleanup fires even when the test fails.
func TestCrossRepoEdges_WriterPersistsEdges(t *testing.T) {
	pgStore := getTestPostgresStore(t)
	ctx := context.Background()

	tenantID := uuid.New()
	tenantIDStr := tenantID.String()

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		// symbol_cache is keyed by tenant_id only and does not cascade
		// from workspaces or repositories.
		if _, err := pgStore.Pool().Exec(cleanupCtx,
			`DELETE FROM symbol_cache WHERE tenant_id = $1`, tenantID); err != nil {
			t.Logf("cleanup: symbol_cache delete failed: %v", err)
		}
		// repositories cascades to entities (ON DELETE CASCADE), which
		// in turn cascades to claims, claim_verifications,
		// cross_repo_edges, and runtime_observations.
		if _, err := pgStore.Pool().Exec(cleanupCtx,
			`DELETE FROM repositories WHERE tenant_id = $1`, tenantID); err != nil {
			t.Logf("cleanup: repositories delete failed: %v", err)
		}
		// workspaces cascades to workspace_modules, cross_module_edges,
		// and runtime_observations.
		if _, err := pgStore.Pool().Exec(cleanupCtx,
			`DELETE FROM workspaces WHERE tenant_id = $1`, tenantID); err != nil {
			t.Logf("cleanup: workspaces delete failed: %v", err)
		}
	})

	ws, err := pgStore.CreateWorkspace(ctx, tenantIDStr, "cross-repo-writer-test", "")
	if err != nil {
		t.Fatalf("CreateWorkspace failed: %v", err)
	}

	repoA, err := pgStore.AddRepository(ctx, ws.ID,
		"local", "test://repo-a", "main", "go", "example.com/corp/auth")
	if err != nil {
		t.Fatalf("AddRepository (A) failed: %v", err)
	}
	repoB, err := pgStore.AddRepository(ctx, ws.ID,
		"local", "test://repo-b", "main", "go", "example.com/corp/gateway")
	if err != nil {
		t.Fatalf("AddRepository (B) failed: %v", err)
	}

	// --- Repo B: the import target. ---
	//
	// Runs first so its package entity exists in the database when
	// detectCrossRepoImports runs for A. Its package_path, package, and
	// name are all set to the module path so findPackageEntityInRepo
	// resolves it via any of the three OR'd columns.
	analysisB := uuid.New()
	resultB := &analyzer.Result{
		Entities: []analyzer.Entity{
			{
				Kind:        analyzer.KindPackage,
				Name:        "example.com/corp/gateway",
				Package:     "example.com/corp/gateway",
				PackagePath: "example.com/corp/gateway",
				ModulePath:  "example.com/corp/gateway",
				File:        "/tmp/fake/gateway/gateway.go",
				Line:        1,
				LineStart:   1,
				LineEnd:     1,
				Exported:    true,
				Language:    "go",
			},
		},
	}
	if err := pgStore.SaveSemanticGraph(ctx, tenantID, ws.ID, repoB.ID,
		analysisB, resultB, "test-commit-b"); err != nil {
		t.Fatalf("SaveSemanticGraph (B) failed: %v", err)
	}

	// --- Repo A: the importer. ---
	//
	// Two entities: a file entity whose file_path matches the IMPORTS
	// relationship's Evidence.File (so fileEntityMap finds it directly),
	// and a package entity for symmetry with real analyzer output.
	//
	// The relationship is a single synthetic IMPORTS edge pointing at
	// repo B's module path. Its From value is cosmetic —
	// detectCrossRepoImports resolves from_entity_id from fileEntityMap,
	// not from rel.From.
	analysisA := uuid.New()
	resultA := &analyzer.Result{
		Entities: []analyzer.Entity{
			{
				Kind:        analyzer.KindFile,
				Name:        "main.go",
				Package:     "example.com/corp/auth",
				PackagePath: "example.com/corp/auth",
				ModulePath:  "example.com/corp/auth",
				File:        "/tmp/fake/auth/main.go",
				Line:        1,
				LineStart:   1,
				LineEnd:     10,
				Exported:    true,
				Language:    "go",
			},
			{
				Kind:        analyzer.KindPackage,
				Name:        "example.com/corp/auth",
				Package:     "example.com/corp/auth",
				PackagePath: "example.com/corp/auth",
				ModulePath:  "example.com/corp/auth",
				File:        "/tmp/fake/auth/main.go",
				Line:        1,
				LineStart:   1,
				LineEnd:     1,
				Exported:    true,
				Language:    "go",
			},
		},
		Relationships: []analyzer.Relationship{
			{
				From:       "main",
				To:         "example.com/corp/gateway",
				Type:       string(analyzer.RelImports),
				Confidence: 1.0,
				Evidence: analyzer.Evidence{
					File:     "/tmp/fake/auth/main.go",
					Line:     3,
					Analyzer: "test/synthetic",
				},
			},
		},
	}
	if err := pgStore.SaveSemanticGraph(ctx, tenantID, ws.ID, repoA.ID,
		analysisA, resultA, "test-commit-a"); err != nil {
		t.Fatalf("SaveSemanticGraph (A) failed: %v", err)
	}

	// --- Assertion: the resolved branch was reached. ---
	//
	// SaveSemanticGraph logs a warning and returns nil when cross-repo
	// detection fails; it does not surface the error to the caller. A
	// row count of 0 is the observable signal that the writer did not
	// land. If this fires, grep the test output for
	// "Failed to insert cross-repo edge" or "Cross-repo detection
	// failed" to see which stage dropped the row.
	var resolvedCount int
	if err := pgStore.Pool().QueryRow(ctx,
		`SELECT COUNT(*) FROM cross_repo_edges
		  WHERE workspace_id = $1 AND resolved = true`,
		ws.ID).Scan(&resolvedCount); err != nil {
		t.Fatalf("count cross_repo_edges failed: %v", err)
	}
	if resolvedCount == 0 {
		t.Fatalf(
			"expected at least 1 resolved cross_repo_edge for workspace %s "+
				"(repoA=%s -> repoB=%s); got 0. Check test output for "+
				"'Failed to insert cross-repo edge' or 'Cross-repo detection failed' "+
				"warnings from SaveSemanticGraph.",
			ws.ID, repoA.ID, repoB.ID)
	}

	// Secondary assertion: the edge's from_entity_id must be the file
	// entity, and to_entity_id must be the target package entity. Both
	// are non-NULL only on the resolved branch.
	var fromNotNull, toNotNull int
	if err := pgStore.Pool().QueryRow(ctx,
		`SELECT
		    COUNT(*) FILTER (WHERE from_entity_id IS NOT NULL),
		    COUNT(*) FILTER (WHERE to_entity_id IS NOT NULL)
		  FROM cross_repo_edges
		  WHERE workspace_id = $1`,
		ws.ID).Scan(&fromNotNull, &toNotNull); err != nil {
		t.Fatalf("count entity FKs failed: %v", err)
	}
	if fromNotNull == 0 {
		t.Fatalf("cross_repo_edges row landed but from_entity_id is NULL; " +
			"detectCrossRepoImports did not resolve the file entity")
	}
	if toNotNull == 0 {
		t.Fatalf("cross_repo_edges row landed but to_entity_id is NULL; " +
			"findPackageEntityInRepo did not resolve repo B's package entity")
	}
}
