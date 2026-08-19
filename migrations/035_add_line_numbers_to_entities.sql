-- 035_add_line_numbers_to_entities.sql
-- Add line number columns to entities and evidence tracking

-- 1. Ensure all candidate line columns exist on entities
ALTER TABLE entities 
ADD COLUMN IF NOT EXISTS line INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS line_start INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS line_end INTEGER DEFAULT 0;

-- 2. Indexes for span queries
CREATE INDEX IF NOT EXISTS idx_entities_line_start ON entities(line_start);
CREATE INDEX IF NOT EXISTS idx_entities_line_end ON entities(line_end);

-- 3. Safely backfill if legacy line column has values
UPDATE entities 
SET line_start = line 
WHERE (line_start = 0 OR line_start IS NULL) AND line IS NOT NULL AND line > 0;

UPDATE entities 
SET line_end = line 
WHERE (line_end = 0 OR line_end IS NULL) AND line IS NOT NULL AND line > 0;

-- 4. Ensure evidence_store table exists and add span columns
CREATE TABLE IF NOT EXISTS evidence_store (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE evidence_store 
ADD COLUMN IF NOT EXISTS line_start INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS line_end INTEGER DEFAULT 0;