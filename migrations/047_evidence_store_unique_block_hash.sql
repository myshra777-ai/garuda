-- Migration 047: Ensure unique constraint on evidence_store.block_hash for ON CONFLICT resolution

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint c
        JOIN pg_class t ON c.conrelid = t.oid
        WHERE t.relname = 'evidence_store' AND c.contype IN ('p', 'u')
    ) THEN
        ALTER TABLE evidence_store ADD CONSTRAINT uq_evidence_store_block_hash UNIQUE (block_hash);
    END IF;
EXCEPTION
    WHEN duplicate_table OR duplicate_object THEN NULL;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS idx_evidence_store_block_hash ON evidence_store (block_hash);