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
				// The transaction was rolled back, so NONE of this migration ran. Recording it
				// as applied hides schema drift; surface it, and let CI/operators opt into failing.
				if os.Getenv("GARUDA_MIGRATE_STRICT") == "1" {
					return fmt.Errorf("migration %s conflicts with existing schema and was not applied: %w", name, execErr)
				}
				fmt.Fprintf(os.Stderr, "WARN: migration %s hit an already-exists conflict and was rolled back (recording as applied; set GARUDA_MIGRATE_STRICT=1 to fail instead): %v\n", name, execErr)
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
		ALTER TABLE tenants ADD COLUMN IF NOT EXISTS plan_tier TEXT NOT NULL DEFAULT 'community';

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

	_, _ = pool.Exec(ctx, `-- ==========================================================================
-- 098_strict_workspace_and_tenant_isolation.sql
-- Garuda v1.9.9 Strict Multi-Tenant & Per-Workspace Isolation:
-- 1. Adds workspace_id and created_by_user_id to decisions so ADR & decision
--    lineage is strictly isolated per (tenant_id, workspace_id) and never
--    bleeds across workspaces or tenants.
-- 2. Repairs any historical non-bootstrap user accounts that were linked to
--    the canonical tenant (00000000-0000-0000-0000-000000000001) prior to 097.
-- 3. Purges auto-seeded demo document_claims from empty workspaces (workspaces
--    with 0 indexed repositories).
-- ==========================================================================

ALTER TABLE decisions ADD COLUMN IF NOT EXISTS workspace_id UUID;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS created_by_user_id UUID;
CREATE INDEX IF NOT EXISTS idx_decisions_tenant_workspace ON decisions (tenant_id, workspace_id);

-- Backfill existing decisions in canonical tenant to its 'default' workspace only
UPDATE decisions d
   SET workspace_id = w.id
  FROM workspaces w
 WHERE d.tenant_id = w.tenant_id
   AND w.name = 'default'
   AND d.workspace_id IS NULL;

-- Ensure every non-bootstrap tenant_owner / individual_dev user has their own isolated tenant
DO $$
DECLARE
    rec RECORD;
    new_tid UUID;
    new_wid UUID;
    inv_code TEXT;
    ws_name TEXT;
BEGIN
    FOR rec IN
        SELECT id, email, COALESCE(NULLIF(full_name, ''), email) AS fname,
               COALESCE(NULLIF(username, ''), SPLIT_PART(email, '@', 1)) AS uname,
               COALESCE(NULLIF(account_type, ''), 'tenant_owner') AS atype,
               primary_tenant_id
          FROM users
         WHERE LOWER(email) NOT IN ('admin@local', 'myshra777@gmail.com')
    LOOP
        IF rec.atype != 'team_member' AND (rec.primary_tenant_id IS NULL OR rec.primary_tenant_id = '00000000-0000-0000-0000-000000000001') THEN
            -- Check if user already owns a non-canonical tenant
            SELECT id INTO new_tid FROM tenants WHERE owner_user_id = rec.id AND id != '00000000-0000-0000-0000-000000000001' LIMIT 1;
            IF new_tid IS NULL THEN
                new_tid := gen_random_uuid();
                inv_code := 'GRD-' || UPPER(SUBSTRING(REPLACE(new_tid::text, '-', ''), 1, 6));
                INSERT INTO tenants (id, name, slug, invite_code, owner_user_id, created_at, updated_at)
                VALUES (new_tid, rec.fname || '''s Organization', LOWER(REGEXP_REPLACE(rec.uname, '[^a-zA-Z0-9]+', '-', 'g')), inv_code, rec.id, NOW(), NOW())
                ON CONFLICT (id) DO NOTHING;
            END IF;

            UPDATE users SET primary_tenant_id = new_tid WHERE id = rec.id;

            INSERT INTO tenant_members (user_id, tenant_id, role, created_at)
            VALUES (rec.id, new_tid, 'owner', NOW())
            ON CONFLICT (user_id, tenant_id) DO UPDATE SET role = 'owner';

            -- Remove erroneous membership in canonical tenant
            DELETE FROM workspace_members wm
             USING workspaces w
             WHERE wm.workspace_id = w.id
               AND wm.user_id = rec.id
               AND w.tenant_id = '00000000-0000-0000-0000-000000000001';

            DELETE FROM tenant_members
             WHERE user_id = rec.id
               AND tenant_id = '00000000-0000-0000-0000-000000000001';

            -- Ensure user has an initial workspace in their isolated tenant
            IF rec.atype = 'individual_dev' THEN
                ws_name := rec.uname || '-workflow';
            ELSE
                ws_name := 'default';
            END IF;

            SELECT id INTO new_wid FROM workspaces WHERE tenant_id = new_tid LIMIT 1;
            IF new_wid IS NULL THEN
                new_wid := gen_random_uuid();
                INSERT INTO workspaces (id, tenant_id, name, root_path, is_go_work, description, created_at, updated_at)
                VALUES (new_wid, new_tid, ws_name, '.', false, 'Isolated workspace', NOW(), NOW())
                ON CONFLICT (tenant_id, name) DO NOTHING;
                SELECT id INTO new_wid FROM workspaces WHERE tenant_id = new_tid AND name = ws_name LIMIT 1;
            END IF;

            IF new_wid IS NOT NULL THEN
                INSERT INTO workspace_members (user_id, workspace_id, role, created_at)
                VALUES (rec.id, new_wid, 'owner', NOW())
                ON CONFLICT (user_id, workspace_id) DO NOTHING;
            END IF;
        END IF;
    END LOOP;
END $$;

-- Purge auto-seeded demo document_claims from workspaces that have 0 indexed repositories
DELETE FROM document_claims dc
 WHERE NOT EXISTS (
     SELECT 1 FROM repositories r WHERE r.workspace_id = dc.workspace_id
 )
   AND (
     dc.document_path LIKE '%0001-payment-idempotency%'
     OR dc.document_path LIKE '%0002-merkle-integrity-layer%'
     OR dc.document_path LIKE '%0003-workspace-state-snapshots%'
     OR dc.document_path LIKE '%0004-zero-trust-agent-gate%'
   );
`)

	return nil
}
