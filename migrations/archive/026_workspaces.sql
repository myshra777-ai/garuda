-- Drop existing tables to reset (ONLY in dev environment)
DROP TABLE IF EXISTS evidence_store CASCADE;
DROP TABLE IF EXISTS decision_revisions CASCADE;
DROP TABLE IF EXISTS decisions CASCADE;
DROP TABLE IF EXISTS merkle_roots CASCADE;
DROP TABLE IF EXISTS repositories CASCADE;
DROP TABLE IF EXISTS workspaces CASCADE;

-- Workspaces (Tenant-scoped)
CREATE TABLE workspaces (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE(tenant_id, name)
);

-- Repositories (Scoped to Workspace)
CREATE TABLE repositories (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    url TEXT NOT NULL,
    default_branch TEXT NOT NULL DEFAULT 'main',
    language TEXT,
    current_commit TEXT,
    enabled BOOLEAN NOT NULL DEFAULT true,
    analysis_status TEXT NOT NULL DEFAULT 'pending',
    last_analyzed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE(workspace_id, url)
);

-- Decisions (Parent entity, Multi-tenant)
CREATE TABLE decisions (
    id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    title TEXT NOT NULL,
    statement TEXT,
    status TEXT NOT NULL,
    domain TEXT NOT NULL,
    system TEXT NOT NULL,
    owner TEXT NOT NULL,
    fingerprint TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, id)
);

-- Decision Revisions (Immutable chain, Multi-tenant)
CREATE TABLE decision_revisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    decision_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    revision_number INT NOT NULL,
    canonical_json JSONB NOT NULL,
    decision_hash BYTEA NOT NULL,
    previous_revision_hash BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (tenant_id, decision_id) REFERENCES decisions(tenant_id, id) ON DELETE CASCADE,
    UNIQUE(tenant_id, decision_id, revision_number)
);

-- Evidence Store (Multi-tenant, content-addressable)
CREATE TABLE evidence_store (
    tenant_id UUID NOT NULL,
    block_hash BYTEA NOT NULL,
    content JSONB NOT NULL,  -- Store structured evidence, not a string!
    ref_count INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, block_hash)
);

-- Merkle Roots (Per-tenant)
CREATE TABLE merkle_roots (
    tenant_id UUID PRIMARY KEY,
    root_hash BYTEA NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- Indexes for performance
CREATE INDEX idx_revisions_decision ON decision_revisions(tenant_id, decision_id, revision_number);
CREATE INDEX idx_decisions_tenant ON decisions(tenant_id);
CREATE INDEX idx_evidence_tenant ON evidence_store(tenant_id);
CREATE INDEX idx_repositories_workspace ON repositories(workspace_id);
CREATE INDEX idx_repositories_status ON repositories(analysis_status);