-- 027_hardening.sql
-- Enforces Multi-Tenant Isolation, Provenance, and Workspace Registry

-- 1. Isolated Content-Addressed Evidence Store
CREATE TABLE IF NOT EXISTS evidence_store (
    tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001'::uuid,
    block_hash BYTEA NOT NULL,
    content JSONB NOT NULL,
    ref_count INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, block_hash)
);

-- 2. Workspace Boundary Tables
CREATE TABLE IF NOT EXISTS workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_workspaces_tenant_name UNIQUE (tenant_id, name)
);

CREATE TABLE IF NOT EXISTS repositories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    provider TEXT NOT NULL DEFAULT 'github',
    url TEXT NOT NULL,
    default_branch TEXT NOT NULL DEFAULT 'main',
    language TEXT DEFAULT 'go',
    current_commit TEXT DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT true,
    analysis_status TEXT NOT NULL DEFAULT 'pending',
    last_analyzed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_repositories_workspace_url UNIQUE (workspace_id, url)
);

-- 3. Hardened Decisions & Revisions with Provenance Links
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS workspace_id UUID REFERENCES workspaces(id) ON DELETE SET NULL;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS repository_id UUID REFERENCES repositories(id) ON DELETE SET NULL;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS commit_sha TEXT DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_decisions_provenance ON decisions(tenant_id, workspace_id, repository_id);
CREATE INDEX IF NOT EXISTS idx_repositories_status ON repositories(tenant_id, analysis_status);