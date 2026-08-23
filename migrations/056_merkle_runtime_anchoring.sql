-- migrations/056_merkle_runtime_anchoring.sql
-- Cryptographic anchoring for runtime observations and claim verifications

ALTER TABLE merkle_snapshots 
    ADD COLUMN IF NOT EXISTS static_root_hash VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS runtime_root_hash VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS runtime_leaf_count BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS verified_claims_count BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS contradicted_claims_count BIGINT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_merkle_snapshots_runtime ON merkle_snapshots(tenant_id, runtime_root_hash);
