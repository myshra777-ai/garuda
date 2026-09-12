-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- 075_document_claims_workspace_id.sql
--
-- Adds workspace_id to document_claims and makes it authoritative.
-- The workspace (string) column was written when workspaces did not
-- yet exist as a table (migration 059 set DEFAULT 'default'). Readers
-- that join on the string silently return zero after a workspace is
-- renamed, because the name in document_claims becomes stale while the
-- workspace row keeps its stable UUID.
--
-- This migration adds the UUID column, backfills it, deletes rows whose
-- workspace string has never corresponded to a real workspace, adds the
-- FK and NOT NULL, and leaves the string column in place for the readers
-- that still use it. Item 3b migrates those readers and drops the string
-- column.
--
-- Entire migration runs in one transaction. If any guard raises, nothing
-- is applied.

BEGIN;

-- Survey: how many rows can be backfilled, how many cannot.
DO $$
DECLARE
  total INT;
  backfillable INT;
  orphans INT;
  r RECORD;
BEGIN
  SELECT COUNT(*) INTO total FROM document_claims;
  SELECT COUNT(*) INTO backfillable
    FROM document_claims dc
    JOIN workspaces w ON w.name = dc.workspace AND w.tenant_id = dc.tenant_id;
  orphans := total - backfillable;

  RAISE NOTICE '075: % total rows, % backfillable, % orphaned',
    total, backfillable, orphans;

  IF orphans > 0 THEN
    RAISE NOTICE '075: orphan workspace strings (no matching workspaces row):';
    FOR r IN
      SELECT DISTINCT dc.workspace, COUNT(*) AS n
        FROM document_claims dc
        LEFT JOIN workspaces w ON w.name = dc.workspace AND w.tenant_id = dc.tenant_id
       WHERE w.id IS NULL
       GROUP BY dc.workspace
    LOOP
      RAISE NOTICE '  workspace=% rows=%', r.workspace, r.n;
    END LOOP;
  END IF;
END $$;

-- Delete rows whose workspace string is the column default from
-- migration 059 and has never corresponded to a real workspace. These
-- are unreachable: no query that filters by workspace name can find
-- them, because there is no workspace named 'default'.
DELETE FROM document_claims
 WHERE workspace = 'default'
   AND NOT EXISTS (
     SELECT 1 FROM workspaces w
      WHERE w.name = document_claims.workspace
        AND w.tenant_id = document_claims.tenant_id
   );

-- Any remaining orphans are a data-integrity decision for the operator.
-- Abort rather than leave them NULL after the NOT NULL is applied.
DO $$
DECLARE
  remaining INT;
  r RECORD;
BEGIN
  SELECT COUNT(*) INTO remaining
    FROM document_claims dc
    LEFT JOIN workspaces w ON w.name = dc.workspace AND w.tenant_id = dc.tenant_id
   WHERE w.id IS NULL;

  IF remaining > 0 THEN
    RAISE NOTICE '075: % rows remain orphaned; list follows', remaining;
    FOR r IN
      SELECT DISTINCT dc.workspace, COUNT(*) AS n
        FROM document_claims dc
        LEFT JOIN workspaces w ON w.name = dc.workspace AND w.tenant_id = dc.tenant_id
       WHERE w.id IS NULL
       GROUP BY dc.workspace
    LOOP
      RAISE NOTICE '  workspace=% rows=%', r.workspace, r.n;
    END LOOP;
    RAISE EXCEPTION '075: reconcile the orphan workspaces above before applying this migration';
  END IF;
END $$;

-- Add the column.
ALTER TABLE document_claims
  ADD COLUMN IF NOT EXISTS workspace_id UUID;

-- Backfill.
UPDATE document_claims dc
   SET workspace_id = w.id
  FROM workspaces w
 WHERE w.name = dc.workspace
   AND w.tenant_id = dc.tenant_id
   AND dc.workspace_id IS NULL;

-- Apply constraints.
ALTER TABLE document_claims
  ALTER COLUMN workspace_id SET NOT NULL;

ALTER TABLE document_claims
  ADD CONSTRAINT document_claims_workspace_id_fkey
  FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_document_claims_workspace_id
  ON document_claims (workspace_id);

COMMIT;