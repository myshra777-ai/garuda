-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- Consolidates duplicate repository rows created by two naming conventions:
--   * old workspace-analysis path used directory name ("gin")
--   * CLI analyze --save path uses module_path ("github.com/gin-gonic/gin")
--
-- Also fixes workspace drift (chi's data landed in the wrong workspace),
-- and adds a UNIQUE constraint to prevent recurrence.

BEGIN;

-- Step 1: Delete empty duplicates — repos with no entities whose module_path
-- also exists on a populated repo in the same tenant.
DELETE FROM repositories
WHERE id IN (
  SELECT r.id
  FROM repositories r
  WHERE NOT EXISTS (SELECT 1 FROM entities e WHERE e.repository_id = r.id)
    AND EXISTS (
      SELECT 1 FROM repositories r2
      JOIN entities e2 ON e2.repository_id = r2.id
      WHERE r2.tenant_id = r.tenant_id
        AND r2.module_path = r.module_path
        AND r2.module_path IS NOT NULL
        AND r2.module_path != ''
        AND r2.id != r.id
    )
);

-- Step 2: Delete orphan repos — no entities, no twin, no claims. These are
-- empty placeholders that never got analyzed.
DELETE FROM repositories
WHERE id IN (
  SELECT r.id
  FROM repositories r
  WHERE NOT EXISTS (SELECT 1 FROM entities e WHERE e.repository_id = r.id)
    AND NOT EXISTS (SELECT 1 FROM claims c WHERE c.repository_id = r.id)
);

-- Step 3: Move any repo whose workspace_id doesn't match where its entities live.
-- This fixes chi (data in garuda-validation, row duplicated in go-validation-10).
UPDATE repositories r
SET workspace_id = (
  SELECT e.workspace_id
  FROM entities e
  WHERE e.repository_id = r.id
  GROUP BY e.workspace_id
  ORDER BY COUNT(*) DESC
  LIMIT 1
)
WHERE EXISTS (SELECT 1 FROM entities e WHERE e.repository_id = r.id)
  AND r.workspace_id != (
    SELECT e.workspace_id
    FROM entities e
    WHERE e.repository_id = r.id
    GROUP BY e.workspace_id
    ORDER BY COUNT(*) DESC
    LIMIT 1
  );

-- Step 4: Sync analysis_status — a repo with entities is 'synced'.
UPDATE repositories r
SET analysis_status = 'synced',
    last_analyzed_at = COALESCE(r.last_analyzed_at, NOW())
WHERE EXISTS (SELECT 1 FROM entities e WHERE e.repository_id = r.id)
  AND (r.analysis_status != 'synced' OR r.last_analyzed_at IS NULL);

-- Step 5: Prevent recurrence — one repo per (tenant, workspace, module_path).
-- Partial index because module_path can be empty for truly-local folders.
CREATE UNIQUE INDEX IF NOT EXISTS uq_repositories_workspace_module
  ON repositories (tenant_id, workspace_id, module_path)
  WHERE module_path IS NOT NULL AND module_path != '';

COMMIT;