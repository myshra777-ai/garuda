-- Unified Merkle Snapshot Table
CREATE TABLE IF NOT EXISTS merkle_snapshots (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    root_hash TEXT NOT NULL,
    block_height BIGINT NOT NULL,
    parent_snapshot_id UUID,                          -- explicit chain link
    snapshot_hash TEXT NOT NULL,                      -- SHA256(tenant_id:root_hash:block_height:epoch_timestamp)
    epoch_timestamp BIGINT NOT NULL,                  -- Unix timestamp (seconds) for deterministic ordering
    snapshot_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for fast lookups
CREATE INDEX idx_merkle_snapshots_tenant ON merkle_snapshots(tenant_id, snapshot_at DESC);
CREATE INDEX idx_merkle_snapshots_parent ON merkle_snapshots(parent_snapshot_id);
CREATE INDEX idx_merkle_snapshots_hash ON merkle_snapshots(snapshot_hash);

-- Foreign key constraint (optional, ensures parent integrity)
ALTER TABLE merkle_snapshots ADD CONSTRAINT fk_merkle_snapshots_parent
    FOREIGN KEY (parent_snapshot_id) REFERENCES merkle_snapshots(id) ON DELETE SET NULL;

-- Constraint: snapshot_hash is unique per tenant (no two snapshots with same hash)
CREATE UNIQUE INDEX idx_merkle_snapshots_uniq_hash ON merkle_snapshots(tenant_id, snapshot_hash);