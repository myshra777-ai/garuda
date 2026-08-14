-- 024_immutable_revisions.sql
-- Step 1: Add tenant_id to decisions (idempotent)
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS tenant_id UUID;

-- Step 2: Ensure decisions table has the right columns
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS fingerprint TEXT;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS evidence_ids BYTEA[];
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS temporal_metadata JSONB;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS scope JSONB;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ;

-- Step 3: Create decision_revisions table
CREATE TABLE IF NOT EXISTS decision_revisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
    decision_id UUID NOT NULL,
    revision_number INT NOT NULL,
    canonical_json JSONB NOT NULL,
    decision_hash BYTEA NOT NULL,
    previous_revision_hash BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_revisions_tenant_decision_rev UNIQUE(tenant_id, decision_id, revision_number),
    CONSTRAINT fk_revisions_decisions FOREIGN KEY (tenant_id, decision_id) REFERENCES decisions(tenant_id, id) ON DELETE CASCADE
);

-- Step 4: Indexes
CREATE INDEX IF NOT EXISTS idx_revisions_decision ON decision_revisions(tenant_id, decision_id, revision_number DESC);
CREATE INDEX IF NOT EXISTS idx_revisions_created ON decision_revisions(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_revisions_hash ON decision_revisions(decision_hash);

-- Step 5: Merkle roots
CREATE TABLE IF NOT EXISTS merkle_roots (
    tenant_id UUID PRIMARY KEY,
    root_hash BYTEA NOT NULL,
    height BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Step 6: Audit events
CREATE TABLE IF NOT EXISTS audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
    event_type TEXT NOT NULL,
    actor TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_tenant ON audit_events(tenant_id, created_at DESC);
