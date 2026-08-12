-- 027_hardening.sql
-- Idempotent schema hardening (multi-tenant, content-addressed evidence, provenance)

-- 1. Add tenant_id to evidence_store (if missing)
ALTER TABLE evidence_store ADD COLUMN IF NOT EXISTS tenant_id UUID;

-- 2. Drop old PK safely (if it exists)
ALTER TABLE evidence_store DROP CONSTRAINT IF EXISTS evidence_store_pkey;

-- 3. Create composite PK
ALTER TABLE evidence_store ADD PRIMARY KEY (tenant_id, block_hash);

-- 4. Add provenance columns to decisions
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS workspace_id UUID;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS repository_id UUID;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS commit_sha TEXT;

-- 5. Indexes
CREATE INDEX IF NOT EXISTS idx_decisions_provenance ON decisions(tenant_id, workspace_id, repository_id);
CREATE INDEX IF NOT EXISTS idx_evidence_tenant ON evidence_store(tenant_id);
