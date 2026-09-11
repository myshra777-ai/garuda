-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- Commit 5b.3b.1 of ADR-0002.
--
-- Adds the columns the v1 decision write path needs, and flips the
-- verification_version defaults on both Merkle tables from 0 to 1.
-- Existing rows retain their recorded version. New rows default to 1.

BEGIN;

-- Decisions: store the v1 proof and label rows explicitly.
ALTER TABLE decisions
  ADD COLUMN IF NOT EXISTS merkle_proof jsonb,
  ADD COLUMN IF NOT EXISTS verification_version smallint NOT NULL DEFAULT 0;

ALTER TABLE decisions
  ADD CONSTRAINT decisions_verification_version_check
  CHECK (verification_version IN (0, 1));

ALTER TABLE decisions
  ALTER COLUMN verification_version SET DEFAULT 1;

-- Merkle tables: new rows default to v1.
ALTER TABLE merkle_roots
  ALTER COLUMN verification_version SET DEFAULT 1;

ALTER TABLE merkle_snapshots
  ALTER COLUMN verification_version SET DEFAULT 1;

COMMIT;