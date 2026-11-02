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
// the schema object already exists.
func isAlreadyAppliedPgError(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "42710", // duplicate_object
			"42P07", // duplicate_table
			"42701", // duplicate_column
			"42P16", // invalid_table_definition
			"23505": // unique_violation
			return true
		}
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "already exists") ||
		strings.Contains(msg, "already applied") ||
		strings.Contains(msg, "multiple primary keys")
}

// Migrate runs all SQL migration files in the migrations directory in sorted order.
func Migrate(connString string, migrationsDir string) error {
	ctx := context.Background()
	migrationsDir = ResolveMigrationsDir(migrationsDir)

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

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

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to list migration files: %w", err)
	}
	sort.Strings(files)

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

	// Always enforce post-migration invariants (B-04 INCONCLUSIVE check + v1.9.9 Multi-Tenant RBAC columns & leak cleanup)
	_, _ = pool.Exec(ctx, `
		ALTER TABLE policy_evaluations DROP CONSTRAINT IF EXISTS policy_evaluations_decision_check;
		ALTER TABLE policy_evaluations ADD CONSTRAINT policy_evaluations_decision_check
			CHECK (decision IN ('ALLOW', 'WARN', 'REVIEW', 'BLOCK', 'INCONCLUSIVE'));

		ALTER TABLE users ADD COLUMN IF NOT EXISTS username TEXT NOT NULL DEFAULT '';
		ALTER TABLE users ADD COLUMN IF NOT EXISTS account_type TEXT NOT NULL DEFAULT 'tenant_owner';
		ALTER TABLE users ADD COLUMN IF NOT EXISTS primary_tenant_id UUID;
		ALTER TABLE users ADD COLUMN IF NOT EXISTS job_title TEXT NOT NULL DEFAULT '';

		ALTER TABLE tenants ADD COLUMN IF NOT EXISTS slug TEXT NOT NULL DEFAULT '';
		ALTER TABLE tenants ADD COLUMN IF NOT EXISTS org_domain TEXT NOT NULL DEFAULT '';
		ALTER TABLE tenants ADD COLUMN IF NOT EXISTS invite_code TEXT NOT NULL DEFAULT '';
		ALTER TABLE tenants ADD COLUMN IF NOT EXISTS owner_user_id UUID;
		ALTER TABLE tenants ADD COLUMN IF NOT EXISTS plan_tier TEXT NOT NULL DEFAULT 'enterprise';

		UPDATE tenants
		   SET invite_code = CASE
		           WHEN id = '00000000-0000-0000-0000-000000000001' THEN 'GRD-GARUDA1'
		           ELSE 'GRD-' || UPPER(SUBSTRING(REPLACE(id::text, '-', ''), 1, 6))
		       END
		 WHERE COALESCE(invite_code, '') = '';

		UPDATE users
		   SET username = LOWER(REGEXP_REPLACE(SPLIT_PART(email, '@', 1), '[^a-zA-Z0-9_]+', '_', 'g'))
		 WHERE COALESCE(username, '') = '';

		DELETE FROM workspace_members wm
		 USING workspaces w, users u
		 WHERE wm.workspace_id = w.id
		   AND wm.user_id = u.id
		   AND w.tenant_id = '00000000-0000-0000-0000-000000000001'
		   AND u.email NOT IN ('admin@local', 'myshra777@gmail.com')
		   AND EXISTS (
		       SELECT 1 FROM tenant_members tm2
		        WHERE tm2.user_id = u.id
		          AND tm2.tenant_id != '00000000-0000-0000-0000-000000000001'
		   );

		DELETE FROM tenant_members tm
		 USING users u
		 WHERE tm.user_id = u.id
		   AND tm.tenant_id = '00000000-0000-0000-0000-000000000001'
		   AND u.email NOT IN ('admin@local', 'myshra777@gmail.com')
		   AND EXISTS (
		       SELECT 1 FROM tenant_members tm2
		        WHERE tm2.user_id = u.id
		          AND tm2.tenant_id != '00000000-0000-0000-0000-000000000001'
		   );
	`)

	return nil
}
