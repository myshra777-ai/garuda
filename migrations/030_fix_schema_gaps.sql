-- 030_fix_schema_gaps.sql
-- Idempotent fix for all schema gaps

-- 1. Fix decision_revisions (add missing columns)
ALTER TABLE decision_revisions ADD COLUMN IF NOT EXISTS tenant_id UUID DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE decision_revisions ADD COLUMN IF NOT EXISTS decision_hash BYTEA;
ALTER TABLE decision_revisions ADD COLUMN IF NOT EXISTS previous_revision_hash BYTEA;
ALTER TABLE decision_revisions ADD COLUMN IF NOT EXISTS canonical_json JSONB;

-- 2. Fix evidence_store (add tenant_id + composite PK)
ALTER TABLE evidence_store ADD COLUMN IF NOT EXISTS tenant_id UUID DEFAULT '00000000-0000-0000-0000-000000000001';

-- Drop old PK if single-column
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints tc
        WHERE tc.table_name = 'evidence_store'
          AND tc.constraint_type = 'PRIMARY KEY'
          AND tc.constraint_name = 'cas_blocks_pkey'
    ) THEN
        ALTER TABLE evidence_store DROP CONSTRAINT cas_blocks_pkey;
    END IF;
END $$;

-- Add composite PK
ALTER TABLE evidence_store ADD PRIMARY KEY (tenant_id, block_hash);

-- 3. Create merkle_roots table (if missing)
CREATE TABLE IF NOT EXISTS merkle_roots (
    tenant_id UUID PRIMARY KEY,
    root_hash BYTEA NOT NULL,
    height BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. Add provenance columns to decisions
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS workspace_id UUID;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS repository_id UUID;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS commit_sha TEXT;

-- 5. Add statement column if missing (alias for title)
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS statement TEXT;

-- 6. Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_decisions_workspace ON decisions(workspace_id);
CREATE INDEX IF NOT EXISTS idx_decisions_repo ON decisions(repository_id);
CREATE INDEX IF NOT EXISTS idx_merkle_roots_tenant ON merkle_roots(tenant_id);
