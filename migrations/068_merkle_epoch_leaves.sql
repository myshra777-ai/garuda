-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- Commit 5b.1 of ADR-0002.
--
-- Stores the ordered leaf list for every Merkle epoch per tenant and
-- per tier. Without this table, an inclusion proof cannot be
-- reconstructed: BuildProof(leaves, index) needs the leaves, and the
-- only place they can live durably is here.
--
-- One row per leaf. An epoch is a (tenant_id, epoch_height) pair.
-- Tier distinguishes the two independent trees ADR-0002 D5 defines:
--   tier 0 — static tree (decision and evaluation hashes)
--   tier 1 — runtime tree (claim verification hashes)
--
-- The table is append-only. No updates, no deletes. A leaf, once
-- recorded at an epoch, is immutable.

BEGIN;

CREATE TABLE IF NOT EXISTS merkle_epoch_leaves (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid        NOT NULL,
    epoch_height    bigint      NOT NULL,
    tier            smallint    NOT NULL,
    leaf_index      int         NOT NULL,
    leaf_hash       text        NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT NOW()
);

ALTER TABLE merkle_epoch_leaves
  ADD CONSTRAINT merkle_epoch_leaves_tier_check
  CHECK (tier IN (0, 1));

ALTER TABLE merkle_epoch_leaves
  ADD CONSTRAINT merkle_epoch_leaves_leaf_hash_format
  CHECK (leaf_hash ~ '^[0-9a-f]{64}$');

ALTER TABLE merkle_epoch_leaves
  ADD CONSTRAINT merkle_epoch_leaves_index_nonneg
  CHECK (leaf_index >= 0);

ALTER TABLE merkle_epoch_leaves
  ADD CONSTRAINT merkle_epoch_leaves_unique_index
  UNIQUE (tenant_id, epoch_height, tier, leaf_index);

CREATE INDEX IF NOT EXISTS idx_merkle_epoch_leaves_lookup
  ON merkle_epoch_leaves(tenant_id, epoch_height, tier, leaf_index);

COMMIT;