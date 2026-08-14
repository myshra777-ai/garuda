-- Add multi-tenant support to decisions table
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS title TEXT;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS fingerprint TEXT;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS evidence_ids BYTEA[];
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS temporal_metadata JSONB;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ;

-- Make tenant_id + id a composite primary key (optional)
-- For now, just add an index
CREATE INDEX IF NOT EXISTS idx_decisions_tenant ON decisions(tenant_id);
