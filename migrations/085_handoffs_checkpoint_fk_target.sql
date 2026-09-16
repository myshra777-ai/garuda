-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- 085_handoffs_checkpoint_fk_target.sql
--
-- Point handoffs.checkpoint_id at agent_checkpoints, the table the
-- handoff code actually writes to.
--
-- Background: two tables hold "checkpoints":
--
--   checkpoints         — id uuid PK, agent_id uuid FK to agents,
--                         state_hash text NOT NULL, restored_at tstz.
--                         No code path inserts into this table.
--
--   agent_checkpoints   — composite PK (tenant_id, id), agent_id text,
--                         checkpoint_name text, status text.
--                         Every read and write in handoff_store.go and
--                         checkpoint_store.go targets this table.
--
-- handoffs.checkpoint_id referenced checkpoints(id) via
-- handoffs_checkpoint_id_fkey. Because no code writes to checkpoints,
-- the FK is unsatisfiable: ExecuteHandoffTransaction inserts into
-- agent_checkpoints (step 7), then inserts into handoffs (step 8)
-- with the agent_checkpoints UUID. The FK check fails with SQLSTATE
-- 23503 on every attempt. The two-agent handoff path has never
-- completed a transaction.
--
-- Fix: repoint the FK at agent_checkpoints. The table has a composite
-- PK (tenant_id, id), so the referencing FK must be composite too.
-- handoffs already carries tenant_id, so the FK becomes
-- (tenant_id, checkpoint_id) -> agent_checkpoints (tenant_id, id).
-- This is strictly stronger than the previous FK: it also prevents a
-- handoff from referencing a checkpoint in a different tenant.
--
-- Preconditions verified 2026-09-16:
--   SELECT COUNT(*) FROM handoffs h
--    WHERE h.checkpoint_id IS NOT NULL
--      AND NOT EXISTS (SELECT 1 FROM agent_checkpoints ac
--                       WHERE ac.id = h.checkpoint_id
--                         AND ac.tenant_id = h.tenant_id);
--   -> 0
--
--   SELECT COUNT(*) FROM checkpoints;
--   -> 0
--
-- The checkpoints table is left in place. Dropping it is a separate
-- decision: no code references it after this migration, but a
-- migration that drops a table cannot be trivially rolled back if a
-- tool or a report is discovered to read it.

BEGIN;

-- Drop the old FK. It is now known to be unsatisfiable.
ALTER TABLE handoffs
  DROP CONSTRAINT IF EXISTS handoffs_checkpoint_id_fkey;

-- Guard: refuse if any row would violate the new composite FK. This
-- is a second check inside the migration, in case the table changed
-- between the precondition query and now.
DO $$
DECLARE
  orphaned bigint;
BEGIN
  SELECT COUNT(*) INTO orphaned
    FROM handoffs h
   WHERE h.checkpoint_id IS NOT NULL
     AND NOT EXISTS (
       SELECT 1 FROM agent_checkpoints ac
        WHERE ac.id = h.checkpoint_id
          AND ac.tenant_id = h.tenant_id
     );
  IF orphaned > 0 THEN
    RAISE EXCEPTION '085: % handoffs rows reference a checkpoint not in agent_checkpoints (tenant_id, id)', orphaned;
  END IF;
END $$;

-- Add the composite FK. ON DELETE RESTRICT preserves the previous
-- behavior: a checkpoint cannot be deleted while a handoff references
-- it.
ALTER TABLE handoffs
  ADD CONSTRAINT handoffs_checkpoint_id_fkey
  FOREIGN KEY (tenant_id, checkpoint_id)
  REFERENCES agent_checkpoints (tenant_id, id)
  ON DELETE RESTRICT;

COMMIT;