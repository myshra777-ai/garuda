-- Migration 049: Align merkle_roots schema with ledger tracker

ALTER TABLE merkle_roots ADD COLUMN IF NOT EXISTS block_height BIGINT DEFAULT 1;
ALTER TABLE merkle_roots ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE merkle_roots ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();