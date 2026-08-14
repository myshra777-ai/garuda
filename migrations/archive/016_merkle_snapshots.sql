-- Migration 016: Unified Production Merkle Root Snapshots Ledger

CREATE TABLE IF NOT EXISTS merkle_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    root_hash TEXT NOT NULL,
    block_height BIGINT NOT NULL,
    parent_snapshot_id UUID,
    snapshot_hash TEXT NOT NULL,
    epoch_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_merkle_snapshots_parent 
        FOREIGN KEY (parent_snapshot_id) REFERENCES merkle_snapshots(id) ON DELETE SET NULL
);

-- Fast lookup indexes
CREATE INDEX IF NOT EXISTS idx_merkle_snapshots_tenant ON merkle_snapshots(tenant_id, epoch_timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_merkle_snapshots_parent ON merkle_snapshots(parent_snapshot_id);
CREATE INDEX IF NOT EXISTS idx_merkle_snapshots_hash ON merkle_snapshots(snapshot_hash);

-- Enforce hash uniqueness per tenant to prevent duplicate identical snapshot entries
CREATE UNIQUE INDEX IF NOT EXISTS idx_merkle_snapshots_uniq_hash ON merkle_snapshots(tenant_id, snapshot_hash);