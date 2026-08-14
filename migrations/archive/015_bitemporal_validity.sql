-- Migration 015: Bitemporal Validity Range

-- Add valid_from and valid_to columns if they don't exist
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS valid_from TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS valid_to TIMESTAMPTZ;

-- Drop constraint if it already exists to prevent SQLSTATE 42710 error
ALTER TABLE decisions DROP CONSTRAINT IF EXISTS check_valid_range;

-- Add valid range check constraint
ALTER TABLE decisions ADD CONSTRAINT check_valid_range CHECK (valid_to IS NULL OR valid_from <= valid_to);

-- Index for fast temporal range filtering
CREATE INDEX IF NOT EXISTS idx_decisions_valid_range ON decisions(tenant_id, valid_from, valid_to);