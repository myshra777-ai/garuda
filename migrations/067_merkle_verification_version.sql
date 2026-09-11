-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- Commit 4 of ADR-0002.
--
-- Adds a `verification_version` discriminator to every Merkle table so
-- that a proof can be verified under the correct scheme.
--
--   version = 0  — written under the deprecated v0 hash chain.
--                  Root hash is SHA256(parent_hex || item_hex) as text.
--                  Not a Merkle tree; no inclusion proof possible.
--
--   version = 1  — written under the canonical encoder + RFC 6962 tree.
--                  See internal/merkle/canonical.go, decision.go, tree.go.
--
-- Existing rows are labeled v0. All new rows default to v1. The migration
-- is additive — no column is dropped, no row is rewritten, no existing
-- proof becomes unverifiable. v0 proofs remain verifiable by the v0 code
-- path that remains in internal/merkle/hash.go (marked Deprecated).
--
-- The migration also adds strict format CHECK constraints to every hash
-- column. All existing values were audited to conform (64 lowercase hex
-- characters). Any future insert of a malformed hash is rejected at the
-- database boundary — this replaces the ad-hoc encode()/decode() calls
-- that were the source of the SQLSTATE 42883 bug in v0.
--
-- Why TEXT with a CHECK constraint rather than BYTEA (revision of ADR-0002 D7):
--   The 11 call sites that read these columns scan into Go `string`
--   fields. Migrating to BYTEA would require touching each of them for no
--   safety gain — a CHECK constraint on TEXT enforces the same invariant
--   (well-formed, 64 chars, lowercase hex) without code churn. The ADR is
--   amended accordingly in the same commit.

BEGIN;

-- ─────────────────────────────────────────────────────────────────
-- 1. Add verification_version to merkle_roots
-- ─────────────────────────────────────────────────────────────────
ALTER TABLE merkle_roots
  ADD COLUMN IF NOT EXISTS verification_version SMALLINT NOT NULL DEFAULT 0;

ALTER TABLE merkle_roots
  ADD CONSTRAINT merkle_roots_version_check
  CHECK (verification_version IN (0, 1));

ALTER TABLE merkle_roots
  ALTER COLUMN verification_version SET DEFAULT 1;

-- ─────────────────────────────────────────────────────────────────
-- 2. Add verification_version to merkle_snapshots
-- ─────────────────────────────────────────────────────────────────
ALTER TABLE merkle_snapshots
  ADD COLUMN IF NOT EXISTS verification_version SMALLINT NOT NULL DEFAULT 0;

ALTER TABLE merkle_snapshots
  ADD CONSTRAINT merkle_snapshots_version_check
  CHECK (verification_version IN (0, 1));

ALTER TABLE merkle_snapshots
  ALTER COLUMN verification_version SET DEFAULT 1;

-- ─────────────────────────────────────────────────────────────────
-- 3. Format CHECK constraints on every hash column
-- ─────────────────────────────────────────────────────────────────

-- merkle_roots
ALTER TABLE merkle_roots
  ADD CONSTRAINT merkle_roots_root_hash_format
  CHECK (root_hash ~ '^[0-9a-f]{64}$');

-- merkle_snapshots: all five hash columns
ALTER TABLE merkle_snapshots
  ADD CONSTRAINT merkle_snapshots_root_hash_format
  CHECK (root_hash ~ '^[0-9a-f]{64}$');

ALTER TABLE merkle_snapshots
  ADD CONSTRAINT merkle_snapshots_snapshot_hash_format
  CHECK (snapshot_hash ~ '^[0-9a-f]{64}$');

ALTER TABLE merkle_snapshots
  ADD CONSTRAINT merkle_snapshots_static_root_hash_format
  CHECK (static_root_hash ~ '^[0-9a-f]{64}$');

ALTER TABLE merkle_snapshots
  ADD CONSTRAINT merkle_snapshots_runtime_root_hash_format
  CHECK (runtime_root_hash ~ '^[0-9a-f]{64}$');

-- ─────────────────────────────────────────────────────────────────
-- 4. Index for cross-version queries
-- ─────────────────────────────────────────────────────────────────
-- Applications that want "only v1 proofs" or "only v0 proofs" can
-- filter on verification_version. The index keeps that fast.
CREATE INDEX IF NOT EXISTS idx_merkle_roots_version
  ON merkle_roots(tenant_id, verification_version);

CREATE INDEX IF NOT EXISTS idx_merkle_snapshots_version
  ON merkle_snapshots(tenant_id, verification_version, block_height DESC);

COMMIT;