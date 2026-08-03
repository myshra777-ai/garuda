-- Extend contradictions table for quarantine and auto-resolution
ALTER TABLE contradictions ADD COLUMN IF NOT EXISTS quarantined BOOLEAN DEFAULT TRUE;
ALTER TABLE contradictions ADD COLUMN IF NOT EXISTS resolution_strategy TEXT DEFAULT 'human';
ALTER TABLE contradictions ADD COLUMN IF NOT EXISTS auto_resolved_at TIMESTAMPTZ;

-- Index for background worker
CREATE INDEX IF NOT EXISTS idx_contradictions_unresolved ON contradictions(tenant_id, resolved, quarantined);