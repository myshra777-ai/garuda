-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- Commit 5b.3b.2 of ADR-0002.
--
-- Adds verification_version to policy_evaluations so that the v0 and
-- v1 verification paths can be distinguished without parsing the
-- proof blob. Existing rows default to 0; new rows default to 1.

BEGIN;

ALTER TABLE policy_evaluations
  ADD COLUMN IF NOT EXISTS verification_version smallint NOT NULL DEFAULT 0;

ALTER TABLE policy_evaluations
  ADD CONSTRAINT policy_evaluations_verification_version_check
  CHECK (verification_version IN (0, 1));

ALTER TABLE policy_evaluations
  ALTER COLUMN verification_version SET DEFAULT 1;

CREATE INDEX IF NOT EXISTS idx_policy_eval_version
  ON policy_evaluations(tenant_id, verification_version);

COMMIT;