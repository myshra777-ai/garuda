-- 032_cross_repo_edges.sql
ALTER TABLE repositories ADD COLUMN IF NOT EXISTS module_path TEXT;

-- Cross-repo edges table
CREATE TABLE IF NOT EXISTS cross_repo_edges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    workspace_id UUID NOT NULL,
    from_repo_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    to_repo_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    from_entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    to_entity_id UUID REFERENCES entities(id) ON DELETE CASCADE, -- nullable for unresolved
    relationship_type TEXT NOT NULL,
    evidence JSONB NOT NULL,
    resolved BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Partial unique index: only enforce uniqueness when to_entity_id is NOT NULL
CREATE UNIQUE INDEX idx_cross_repo_edges_unique_resolved ON cross_repo_edges (
    tenant_id, from_repo_id, to_repo_id, from_entity_id, to_entity_id, relationship_type
) WHERE to_entity_id IS NOT NULL;

-- Allow duplicates when to_entity_id IS NULL (unresolved), but we'll still deduplicate in code
CREATE INDEX idx_cross_repo_edges_tenant ON cross_repo_edges(tenant_id);
CREATE INDEX idx_cross_repo_edges_workspace ON cross_repo_edges(workspace_id);
CREATE INDEX idx_cross_repo_edges_from_repo ON cross_repo_edges(from_repo_id);
CREATE INDEX idx_cross_repo_edges_to_repo ON cross_repo_edges(to_repo_id);
CREATE INDEX idx_cross_repo_edges_from_entity ON cross_repo_edges(from_entity_id);
CREATE INDEX idx_cross_repo_edges_to_entity ON cross_repo_edges(to_entity_id);

-- For unresolved edges, we'll use application-level deduplication