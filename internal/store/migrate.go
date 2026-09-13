// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

// splitGooseUp returns the forward-migration section of a file
// written for the goose migration tool.
//
// Files written for goose have the shape:
//
//	-- +goose Up
//	CREATE TABLE ...
//
//	-- +goose Down
//	DROP TABLE ...
//
// The Down section is a rollback script. Executing it during the
// forward pass would immediately undo what Up just did. That is what
// happened to migrations/003_create_users_table.sql: the runner
// passed the entire file to tx.Exec, Postgres executed CREATE TABLE
// followed by DROP TABLE, the table was gone, and the migration was
// still recorded as applied. The next migration that referenced
// users failed with SQLSTATE 42P01.
//
// The split is on the exact marker line. If the marker is absent,
// the whole file is returned unchanged so migrations that do not
// use goose markers still work.
//
// Migrations that contain a Down section should also have it
// removed at the file level — this helper is defense in depth, not
// an excuse to keep rollback scripts in the forward-only path.
func splitGooseUp(content []byte) []byte {
	const marker = "-- +goose Down"
	if idx := bytes.Index(content, []byte(marker)); idx >= 0 {
		return content[:idx]
	}
	return content
}

// Migrate runs all SQL migration files in the migrations directory in sorted order.
// Each migration file is executed inside an atomic transaction.
func Migrate(connString string, migrationsDir string) error {
	ctx := context.Background()

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

		// Execute the entire migration file inside an atomic transaction
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for %s: %w", name, err)
		}

		if _, err := tx.Exec(ctx, string(splitGooseUp(content))); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("failed to execute migration %s: %w", name, err)
		}

		// Record the migration in the same transaction
		if _, err := tx.Exec(ctx, "INSERT INTO migrations (name) VALUES ($1)", name); err != nil {
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
