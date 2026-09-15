-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- 084_contradictions_runtime_kind.sql
--
-- The contradictions table was created by migrations 005 and 008 for
-- the decision-vs-decision model (FKs on decision_a and decision_b,
-- referencing decisions). Migration 011 added the resolution workflow
-- columns (quarantined, resolution_strategy, auto_resolved_at).
--
-- The verifier's statement 2 (internal/runtime/verifier.go) has been
-- trying to write runtime-vs-static contradictions into this table
-- since the day it was written. Six of its ten INSERT columns do not
-- exist. Every tick has failed silently under an empty
-- "if err != nil { }" block. Verified 2026-09-15: 0 rows, 0 errors
-- surfaced.
--
-- This migration adds a `kind` discriminator and the runtime columns
-- the verifier needs. The decision-vs-decision rows keep their FKs;
-- Postgres composite FKs use MATCH SIMPLE by default, so a NULL
-- decision_a or decision_b satisfies the FK rather than violating it.
-- The existing decision rows are unaffected.
--
-- Both kinds share the resolution workflow columns: quarantined,
-- resolved, resolution_strategy, resolved_at, auto_resolved_at.
-- A runtime contradiction quarantines with resolved = FALSE and is
-- later resolved by a human or by auto-supersede, the same as a
-- decision conflict.
--
-- Run inside one transaction. If any guard raises, nothing applies.

BEGIN;

-- Guard: refuse if a `kind` column already exists. Re-apply should be
-- a no-op, not a silent partial.
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
     WHERE table_name = 'contradictions' AND column_name = 'kind'
  ) THEN
    RAISE EXCEPTION '084: contradictions.kind already exists; migration already applied';
  END IF;
END $$;

-- 1. Discriminator. Existing rows default to decision_vs_decision.
ALTER TABLE contradictions
  ADD COLUMN kind TEXT NOT NULL DEFAULT 'decision_vs_decision';

-- 2. Decision columns become nullable for runtime rows. The composite
-- FKs remain in place; MATCH SIMPLE means a NULL column satisfies the
-- FK without needing to drop and recreate it.
ALTER TABLE contradictions ALTER COLUMN decision_a DROP NOT NULL;
ALTER TABLE contradictions ALTER COLUMN decision_b DROP NOT NULL;

-- 3. Runtime columns.
ALTER TABLE contradictions
  ADD COLUMN workspace_id        UUID,
  ADD COLUMN source_entity_id    UUID,
  ADD COLUMN target_entity_id    UUID,
  ADD COLUMN raw_target          TEXT,
  ADD COLUMN observation_summary TEXT,
  ADD COLUMN evidence_file       TEXT,
  ADD COLUMN evidence_line       INT,
  ADD COLUMN updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- 4. Kind-specific shape constraints.
ALTER TABLE contradictions
  ADD CONSTRAINT chk_contradictions_kind_decision CHECK (
    kind <> 'decision_vs_decision'
    OR (decision_a IS NOT NULL AND decision_b IS NOT NULL)
  );

ALTER TABLE contradictions
  ADD CONSTRAINT chk_contradictions_kind_runtime CHECK (
    kind <> 'runtime_vs_static'
    OR (workspace_id IS NOT NULL
        AND source_entity_id IS NOT NULL
        AND raw_target IS NOT NULL
        AND observation_summary IS NOT NULL)
  );

-- 5. FKs for the runtime rows.
ALTER TABLE contradictions
  ADD CONSTRAINT fk_contradictions_workspace
  FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE;

ALTER TABLE contradictions
  ADD CONSTRAINT fk_contradictions_source_entity
  FOREIGN KEY (source_entity_id) REFERENCES entities(id) ON DELETE CASCADE;

-- 6. Indexes for the runtime query path.
CREATE INDEX idx_contradictions_kind
  ON contradictions (kind);

CREATE INDEX idx_contradictions_runtime_lookup
  ON contradictions (workspace_id, raw_target)
  WHERE kind = 'runtime_vs_static';

COMMIT;