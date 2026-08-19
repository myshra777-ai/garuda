-- 032_cross_repo_edges.sql
-- Adds support for cross-repository dependency detection and graph relationships

-- Ensure repositories table exists before altering or referencing
CREATE TABLE IF NOT EXISTS repositories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    workspace_id UUID,
    name TEXT NOT NULL,
    url TEXT,
    module_path TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT repositories_tenant_name_uniq UNIQUE (tenant_id, name)
);

-- Ensure entities table exists before establishing foreign keys
CREATE TABLE IF NOT EXISTS entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    repository_id UUID REFERENCES repositories(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    kind TEXT NOT NULL,
    package TEXT NOT NULL,
    file TEXT NOT NULL,
    line_start INT,
    line_end INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add module_path column if repositories table was created by an earlier schema
ALTER TABLE repositories ADD COLUMN IF NOT EXISTS module_path TEXT;

-- Cross-repo edges table stores relationships between entities in different repositories
CREATE TABLE IF NOT EXISTS cross_repo_edges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    workspace_id UUID NOT NULL,
    from_repo_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    to_repo_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    from_entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    to_entity_id UUID REFERENCES entities(id) ON DELETE CASCADE, -- NULL for unresolved targets
    relationship_type TEXT NOT NULL, -- 'IMPORTS', 'CALLS', etc.
    evidence JSONB NOT NULL,         -- File, line, commit, analyzer
    resolved BOOLEAN DEFAULT FALSE,  -- True when to_entity_id is resolved
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Performance Indexes
CREATE INDEX IF NOT EXISTS idx_cross_repo_edges_tenant ON cross_repo_edges(tenant_id);
CREATE INDEX IF NOT EXISTS idx_cross_repo_edges_workspace ON cross_repo_edges(workspace_id);
CREATE INDEX IF NOT EXISTS idx_cross_repo_edges_from_repo ON cross_repo_edges(from_repo_id);
CREATE INDEX IF NOT EXISTS idx_cross_repo_edges_to_repo ON cross_repo_edges(to_repo_id);
CREATE INDEX IF NOT EXISTS idx_cross_repo_edges_from_entity ON cross_repo_edges(from_entity_id);
CREATE INDEX IF NOT EXISTS idx_cross_repo_edges_to_entity ON cross_repo_edges(to_entity_id);
CREATE INDEX IF NOT EXISTS idx_repositories_tenant_module ON repositories (tenant_id, module_path);

-- Partial unique index: enforce uniqueness when to_entity_id is resolved
CREATE UNIQUE INDEX IF NOT EXISTS idx_cross_repo_edges_unique_resolved ON cross_repo_edges (
    tenant_id,
    from_repo_id,
    to_repo_id,
    from_entity_id,
    to_entity_id,
    relationship_type
) WHERE to_entity_id IS NOT NULL;

-- Timestamp trigger
CREATE OR REPLACE FUNCTION update_cross_repo_edges_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_cross_repo_edges_updated_at ON cross_repo_edges;
CREATE TRIGGER trigger_cross_repo_edges_updated_at
    BEFORE UPDATE ON cross_repo_edges
    FOR EACH ROW
    EXECUTE FUNCTION update_cross_repo_edges_updated_at();