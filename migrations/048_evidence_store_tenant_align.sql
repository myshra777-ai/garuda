-- Migration 048: Align evidence_store tenant_id constraint and unique index

ALTER TABLE evidence_store ADD COLUMN IF NOT EXISTS tenant_id UUID;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uq_evidence_store_tenant_block_hash'
    ) THEN
        ALTER TABLE evidence_store ADD CONSTRAINT uq_evidence_store_tenant_block_hash UNIQUE (tenant_id, block_hash);
    END IF;
EXCEPTION
    WHEN duplicate_table OR duplicate_object THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_evidence_store_lookup ON evidence_store (tenant_id, block_hash);