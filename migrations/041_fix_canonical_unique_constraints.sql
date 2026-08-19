-- 041_fix_canonical_unique_constraints.sql
-- Harden canonical identity columns, unique constraints, and traversal indexes

-- 1. Drop legacy collision-prone constraints and indexes if present
ALTER TABLE entities DROP CONSTRAINT IF EXISTS entities_tenant_id_workspace_id_repository_id_name_package_key;
DROP INDEX IF EXISTS idx_entities_unique_name_pkg;

-- 2. Ensure all canonical identity columns exist on entities (including workspace_id)
ALTER TABLE entities 
    ADD COLUMN IF NOT EXISTS workspace_id UUID,
    ADD COLUMN IF NOT EXISTS repository_id UUID,
    ADD COLUMN IF NOT EXISTS line INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS module_path TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS package_path TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS receiver_type TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS commit_sha TEXT DEFAULT '';

-- 3. Set up indexes for fast entity lookup and traversal
CREATE INDEX IF NOT EXISTS idx_entities_workspace_id ON entities(workspace_id);
CREATE INDEX IF NOT EXISTS idx_entities_module_path ON entities(module_path);
CREATE INDEX IF NOT EXISTS idx_entities_package_path ON entities(package_path);
CREATE INDEX IF NOT EXISTS idx_entities_receiver_type ON entities(receiver_type);
CREATE INDEX IF NOT EXISTS idx_entities_canonical_lookup ON entities(tenant_id, workspace_id, repository_id, package_path, name);

-- 4. Ensure claims table exists and align schema columns
CREATE TABLE IF NOT EXISTS claims (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE claims 
    ADD COLUMN IF NOT EXISTS workspace_id UUID,
    ADD COLUMN IF NOT EXISTS repository_id UUID,
    ADD COLUMN IF NOT EXISTS from_entity_id UUID,
    ADD COLUMN IF NOT EXISTS to_entity_id UUID,
    ADD COLUMN IF NOT EXISTS epistemic_class TEXT DEFAULT 'OBSERVATION',
    ADD COLUMN IF NOT EXISTS confidence FLOAT DEFAULT 1.0,
    ADD COLUMN IF NOT EXISTS file_path TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS commit_sha TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS line_start INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS line_end INTEGER DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_claims_epistemic_class ON claims(epistemic_class);
CREATE INDEX IF NOT EXISTS idx_claims_traversal ON claims(workspace_id, from_entity_id, to_entity_id);