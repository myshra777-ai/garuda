-- Migration 012: Cryptographic Evidence Chain and Merkle Attestation

CREATE TABLE IF NOT EXISTS merkle_roots (
    tenant_id UUID PRIMARY KEY,
    root_hash TEXT NOT NULL,
    block_height BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add Merkle hash columns to decisions table
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS merkle_hash TEXT;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS parent_merkle_hash TEXT;

-- Evidence blocks table for granular evidence payload hashing
CREATE TABLE IF NOT EXISTS evidence_blocks (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    decision_id UUID NOT NULL,
    prev_hash TEXT NOT NULL,
    evidence_hash TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for fast verification queries
CREATE INDEX IF NOT EXISTS idx_decisions_merkle_hash ON decisions(merkle_hash);
CREATE INDEX IF NOT EXISTS idx_evidence_blocks_decision ON evidence_blocks(tenant_id, decision_id);