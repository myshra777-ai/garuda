-- Drop the old decisions table (CASCADE will remove foreign key constraints automatically)
DROP TABLE IF EXISTS decisions CASCADE;

-- Create new decisions table with composite primary key (tenant_id, id)
CREATE TABLE decisions (
    tenant_id UUID NOT NULL,
    id UUID NOT NULL,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'DRAFT',
    fingerprint TEXT,
    evidence_ids BYTEA[],
    temporal_metadata JSONB,
    scope JSONB,
    owner TEXT,
    confidence FLOAT,
    parent_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMPTZ,
    rationale TEXT,
    alternatives JSONB,
    approvers JSONB,
    dependencies JSONB,
    affected_systems JSONB,
    authorized_by TEXT,
    signature BYTEA,
    PRIMARY KEY (tenant_id, id)
);

-- Add indexes for efficient queries
CREATE INDEX idx_decisions_tenant_status ON decisions(tenant_id, status);
CREATE INDEX idx_decisions_tenant_created ON decisions(tenant_id, created_at DESC);
CREATE INDEX idx_decisions_scope ON decisions USING GIN (scope);
CREATE INDEX idx_decisions_parent ON decisions(parent_id);
