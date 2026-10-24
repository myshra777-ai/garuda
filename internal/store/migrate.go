// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// splitGooseUp returns the forward-migration section of a file
// written for the goose migration tool.
func splitGooseUp(content []byte) []byte {
	const marker = "-- +goose Down"
	if idx := bytes.Index(content, []byte(marker)); idx >= 0 {
		return content[:idx]
	}
	return content
}

// ResolveMigrationsDir locates the migrations directory relative to the
// current working directory or the running binary.
func ResolveMigrationsDir(preferred string) string {
	if preferred != "" {
		if info, err := os.Stat(preferred); err == nil && info.IsDir() {
			return preferred
		}
	}
	candidates := []string{"migrations", "../migrations"}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "migrations"),
			filepath.Join(exeDir, "..", "migrations"),
		)
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return preferred
}

// isAlreadyAppliedPgError returns true when a migration error indicates
// the schema object already exists (for example, when migrations were
// previously applied manually via psql -f outside the migrations table).
func isAlreadyAppliedPgError(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "42710", // duplicate_object (type, constraint, policy already exists)
			"42P07", // duplicate_table (relation already exists)
			"42701", // duplicate_column (column already exists)
			"42P16", // invalid_table_definition (e.g. multiple primary keys on re-run)
			"23505": // unique_violation on idempotent seed insert
			return true
		}
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "already exists") ||
		strings.Contains(msg, "already applied") ||
		strings.Contains(msg, "multiple primary keys")
}

// Migrate runs all SQL migration files in the migrations directory in sorted order.
// Each migration file is executed inside an atomic transaction, and handles
// pre-populated schemas idempotently.
func Migrate(connString string, migrationsDir string) error {
	ctx := context.Background()
	migrationsDir = ResolveMigrationsDir(migrationsDir)

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	// 1. Create tracking table
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// 2. Discover and sort migration files
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to list migration files: %w", err)
	}
	sort.Strings(files)

	// 3. Apply migrations in order
	for _, file := range files {
		name := filepath.Base(file)

		var exists bool
		err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM migrations WHERE name = $1)", name).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration %s: %w", name, err)
		}
		if exists {
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", name, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for %s: %w", name, err)
		}

		if _, execErr := tx.Exec(ctx, string(splitGooseUp(content))); execErr != nil {
			_ = tx.Rollback(ctx)
			if isAlreadyAppliedPgError(execErr) {
				// Object already exists from a manual psql run; record migration and continue.
				_, _ = pool.Exec(ctx, "INSERT INTO migrations (name) VALUES ($1) ON CONFLICT (name) DO NOTHING", name)
				continue
			}
			return fmt.Errorf("failed to execute migration %s: %w", name, execErr)
		}

		if _, err := tx.Exec(ctx, "INSERT INTO migrations (name) VALUES ($1) ON CONFLICT (name) DO NOTHING", name); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("failed to record migration %s: %w", name, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction for %s: %w", name, err)
		}

		fmt.Printf("Applied migration: %s\n", name)
	}

	return nil
}
