// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"os"
	"testing"

	"github.com/myshra777-ai/garuda/internal/store"
)

// getTestPostgresStore returns a *store.PostgresStore connected to the
// test database, or skips the test if no database is configured. This
// mirrors the helper in test/benchmark/tenant_isolation_test.go so both
// suites behave the same way: DB-backed tests run when
// GARUDA_TEST_DATABASE_URL or DATABASE_URL is set, and skip cleanly
// otherwise, so `go test ./...` on a machine without Postgres still
// passes.
//
// Cleanup closes the pool at test end. Tests that mutate data should
// register their own t.Cleanup to delete what they create.
func getTestPostgresStore(t *testing.T) *store.PostgresStore {
	t.Helper()

	dbURL := os.Getenv("GARUDA_TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		t.Skip("skipping DB-backed api test: neither GARUDA_TEST_DATABASE_URL nor DATABASE_URL is set")
	}

	pgStore, err := store.NewPostgresStore(dbURL)
	if err != nil {
		t.Fatalf("failed to initialize postgres store: %v", err)
	}

	t.Cleanup(func() {
		pgStore.Close()
	})

	return pgStore
}
