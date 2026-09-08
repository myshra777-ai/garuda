// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func TestSchema_IndexesAndInvariants(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping database schema validation test")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	requiredIndexes := []string{
		"idx_entities_canonical_name",
		"idx_relationships_src_target",
		"idx_cross_repo_bridges",
	}

	for _, idx := range requiredIndexes {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT 1 FROM pg_indexes 
				WHERE indexname = $1
			);`
		if err := db.QueryRowContext(ctx, query, idx).Scan(&exists); err != nil {
			t.Fatalf("failed checking index %s: %v", idx, err)
		}
		if !exists {
			t.Errorf("critical index missing: %s", idx)
		}
	}
}
