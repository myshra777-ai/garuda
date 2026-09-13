-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- 076_repositories_workspace_scoped_unique.sql
--
-- Replaces the tenant-scoped unique constraint on repositories with a
-- workspace-scoped one.
--
-- Before this migration, repositories was constrained by
-- (tenant_id, name). Two workspaces under the same tenant could not
-- have a repository with the same display name, even when those
-- workspaces were unrelated. The chi collision was the first case
-- where this surfaced: adding github.com/go-chi/chi/v5 to
-- go-validation-10 failed because a row with the same name existed
-- in garuda-validation. Every future case where two workspaces want
-- the same module would have failed identically.
--
-- Every other column and every query on repositories is workspace-
-- scoped. The constraint was the one place the scope did not match.
--
-- Preconditions:
--   1. Zero rows share (tenant_id, workspace_id, name). Verified
--      before this migration was written:
--
--        SELECT tenant_id, workspace_id, name, COUNT(*)
--          FROM repositories
--         GROUP BY tenant_id, workspace_id, name
--        HAVING COUNT(*) > 1;
--        -> (0 rows)
--
--   2. workspace_id is populated for every row. If any row has a
--      NULL workspace_id, that row will not be covered by the new
--      unique constraint (Postgres treats NULLs as distinct in a
--      unique index). A separate migration would be needed to
--      backfill before this one is meaningful.
--
-- Entire migration runs in one transaction. If either the drop or
-- the add fails, nothing is committed.

BEGIN;

-- Guard: refuse if any row has a NULL workspace_id. The new
-- constraint does not constrain those rows, and shipping a
-- constraint that silently does not apply is the class of problem
-- this migration series exists to eliminate.
DO $$
DECLARE
  null_count INT;
BEGIN
  SELECT COUNT(*) INTO null_count FROM repositories WHERE workspace_id IS NULL;
  IF null_count > 0 THEN
    RAISE EXCEPTION '076: % repositories have NULL workspace_id; backfill before applying', null_count;
  END IF;
END $$;

-- Guard: refuse if duplicates already exist under the new key.
DO $$
DECLARE
  dup_count INT;
BEGIN
  SELECT COUNT(*) INTO dup_count FROM (
    SELECT 1 FROM repositories
     GROUP BY tenant_id, workspace_id, name
    HAVING COUNT(*) > 1
  ) d;
  IF dup_count > 0 THEN
    RAISE EXCEPTION '076: % duplicate (tenant, workspace, name) groups exist; reconcile before applying', dup_count;
  END IF;
END $$;

-- Drop the old tenant-only constraint if present.
ALTER TABLE repositories
  DROP CONSTRAINT IF EXISTS repositories_tenant_name_uniq;

-- Drop the target constraint if present to guarantee idempotent re-runs.
ALTER TABLE repositories
  DROP CONSTRAINT IF EXISTS repositories_tenant_workspace_name_uniq;

-- Add the workspace-scoped constraint. The name follows the pattern
-- of repositories_tenant_name_uniq so grep across migrations finds
-- both.
ALTER TABLE repositories
  ADD CONSTRAINT repositories_tenant_workspace_name_uniq
  UNIQUE (tenant_id, workspace_id, name);

COMMIT;