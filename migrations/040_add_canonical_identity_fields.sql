-- 040_add_canonical_identity_fields.sql
-- Add canonical identity fields to entities table and establish claims schema

-- 1. Ensure entities table has canonical identity columns
ALTER TABLE entities 
ADD COLUMN IF NOT EXISTS line INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS module_path TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS package_path TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS receiver_type TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS commit_sha TEXT DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_entities_module_path ON entities(module_path);
CREATE INDEX IF NOT EXISTS idx_entities_package_path ON entities(package_path);
CREATE INDEX IF NOT EXISTS idx_entities_receiver_type ON entities(receiver_type);

-- Safely sync line numbers
UPDATE entities SET line = line_start WHERE line = 0 AND line_start > 0;

-- 2. Ensure claims table exists
CREATE TABLE IF NOT EXISTS claims (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    workspace_id UUID,
    repository_id UUID REFERENCES repositories(id) ON DELETE CASCADE,
    from_entity_id UUID REFERENCES entities(id) ON DELETE CASCADE,
    to_entity_id UUID REFERENCES entities(id) ON DELETE CASCADE,
    subject_entity_id UUID REFERENCES entities(id) ON DELETE CASCADE,
    object_entity_id UUID REFERENCES entities(id) ON DELETE CASCADE,
    claim_type TEXT NOT NULL DEFAULT 'REFERENCES',
    epistemic_class TEXT DEFAULT 'OBSERVED_CODE',
    confidence FLOAT DEFAULT 1.0,
    evidence JSONB DEFAULT '{}',
    snapshot_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Add columns and indexes to claims
ALTER TABLE claims 
ADD COLUMN IF NOT EXISTS epistemic_class TEXT DEFAULT 'OBSERVED_CODE',
ADD COLUMN IF NOT EXISTS confidence FLOAT DEFAULT 1.0;

CREATE INDEX IF NOT EXISTS idx_claims_tenant ON claims(tenant_id);
CREATE INDEX IF NOT EXISTS idx_claims_epistemic_class ON claims(epistemic_class);