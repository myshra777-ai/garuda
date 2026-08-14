-- 024_immutable_revisions.sql

-- 1. Ensure decisions table exists
CREATE TABLE IF NOT EXISTS decisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    statement TEXT,
    status TEXT NOT NULL DEFAULT 'PROPOSED',
    domain TEXT DEFAULT 'general',
    system TEXT DEFAULT 'cli',
    team TEXT DEFAULT '',
    env TEXT DEFAULT '',
    owner TEXT DEFAULT 'system',
    confidence DOUBLE PRECISION DEFAULT 1.0,
    fingerprint TEXT DEFAULT '',
    parent_id UUID,
    temporal_metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMPTZ
);

-- 2. Add tenant_id column if missing
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001';

-- 3. Create decision_revisions table
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
    CONSTRAINT fk_revisions_decisions FOREIGN KEY (decision_id) REFERENCES decisions(id) ON DELETE CASCADE
);

-- 4. Indexes
CREATE INDEX IF NOT EXISTS idx_revisions_decision ON decision_revisions(tenant_id, decision_id, revision_number DESC);
CREATE INDEX IF NOT EXISTS idx_revisions_created ON decision_revisions(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_revisions_hash ON decision_revisions(decision_hash);

-- 5. Current decisions view
CREATE OR REPLACE VIEW current_decisions AS
SELECT DISTINCT ON (r.tenant_id, r.decision_id)
    r.decision_id AS decision_id,
    r.tenant_id AS tenant_id,
    r.created_at AS decision_created_at,
    r.revision_number,
    r.canonical_json,
    r.decision_hash,
    r.previous_revision_hash,
    r.created_at AS revision_created_at
FROM decision_revisions r
ORDER BY r.tenant_id, r.decision_id, r.revision_number DESC;

-- 6. Merkle roots
CREATE TABLE IF NOT EXISTS merkle_roots (
    tenant_id UUID PRIMARY KEY,
    root_hash BYTEA NOT NULL,
    height BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 7. Audit events compatibility
CREATE TABLE IF NOT EXISTS audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
    event_type TEXT NOT NULL DEFAULT 'DECISION_STATE_CHANGE',
    actor TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Ensure tenant_id and created_at exist on legacy audit_events instances
ALTER TABLE audit_events ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE audit_events ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Index using tenant_id and created_at
CREATE INDEX IF NOT EXISTS idx_audit_tenant ON audit_events(tenant_id, created_at DESC);