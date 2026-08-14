-- Migration 013: Bitemporal Validity Range Columns & Constraints

-- 1. Ensure scope JSONB column exists
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS scope JSONB DEFAULT '{}'::jsonb;

-- 2. Add validity columns
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS valid_from TIMESTAMPTZ;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS valid_to TIMESTAMPTZ;

-- 3. Backfill existing decisions: valid_from = created_at
UPDATE decisions SET valid_from = created_at WHERE valid_from IS NULL;

-- 4. Set valid_from to NOT NULL after backfill
ALTER TABLE decisions ALTER COLUMN valid_from SET NOT NULL;

-- 5. Add indexes for bitemporal point-in-time and range queries
CREATE INDEX IF NOT EXISTS idx_decisions_valid_range ON decisions(tenant_id, valid_from, valid_to);

-- 6. Safely enforce valid_range constraint without PL/pgSQL DO blocks
ALTER TABLE decisions DROP CONSTRAINT IF EXISTS check_valid_range;
ALTER TABLE decisions ADD CONSTRAINT check_valid_range CHECK (valid_to IS NULL OR valid_from <= valid_to);