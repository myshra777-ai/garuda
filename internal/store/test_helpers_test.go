// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"os"
	"testing"
)

// getTestPostgresStore returns a *PostgresStore connected to the test
// database, or skips the test if no database is configured. Mirrors
// the same helper in internal/api and test/benchmark.
func getTestPostgresStore(t *testing.T) *PostgresStore {
	t.Helper()

	dbURL := os.Getenv("GARUDA_TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		t.Skip("skipping DB-backed store test: neither GARUDA_TEST_DATABASE_URL nor DATABASE_URL is set")
	}

	pgStore, err := NewPostgresStore(dbURL)
	if err != nil {
		t.Fatalf("failed to initialize postgres store: %v", err)
	}

	t.Cleanup(func() {
		pgStore.Close()
	})

	return pgStore
}
