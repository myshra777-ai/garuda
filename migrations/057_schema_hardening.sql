-- =========================================================================
-- 057_schema_hardening.sql: Production Parity, Isolation & Indexes
-- =========================================================================

-- 1. Workspace and Repository Alignment
ALTER TABLE workspaces ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE workspaces ALTER COLUMN root_path DROP NOT NULL;
ALTER TABLE workspaces ALTER COLUMN root_path SET DEFAULT '.';
ALTER TABLE workspaces ADD COLUMN IF NOT EXISTS is_go_work BOOLEAN DEFAULT false;

ALTER TABLE repositories ADD COLUMN IF NOT EXISTS name TEXT DEFAULT 'primary';
ALTER TABLE repositories ALTER COLUMN name DROP NOT NULL;
ALTER TABLE repositories ALTER COLUMN tenant_id DROP NOT NULL;
ALTER TABLE repositories ADD COLUMN IF NOT EXISTS provider TEXT DEFAULT 'local';
ALTER TABLE repositories ADD COLUMN IF NOT EXISTS url TEXT;
ALTER TABLE repositories ADD COLUMN IF NOT EXISTS default_branch TEXT DEFAULT 'main';
ALTER TABLE repositories ADD COLUMN IF NOT EXISTS language TEXT DEFAULT 'go';
ALTER TABLE repositories ADD COLUMN IF NOT EXISTS module_path TEXT DEFAULT '';
ALTER TABLE repositories ADD COLUMN IF NOT EXISTS enabled BOOLEAN DEFAULT true;
ALTER TABLE repositories ADD COLUMN IF NOT EXISTS analysis_status TEXT DEFAULT 'pending';
ALTER TABLE repositories ADD COLUMN IF NOT EXISTS current_commit TEXT;
ALTER TABLE repositories ADD COLUMN IF NOT EXISTS last_analyzed_at TIMESTAMPTZ;
CREATE UNIQUE INDEX IF NOT EXISTS uq_repositories_workspace_url ON repositories (workspace_id, url);

-- 2. Entities Canonical Identity & AST Fields
ALTER TABLE entities ADD COLUMN IF NOT EXISTS canonical_name TEXT;
ALTER TABLE entities ADD COLUMN IF NOT EXISTS analysis_id UUID;
ALTER TABLE entities ADD COLUMN IF NOT EXISTS file_path TEXT;
ALTER TABLE entities ALTER COLUMN file DROP NOT NULL;
ALTER TABLE entities ALTER COLUMN file SET DEFAULT '';
ALTER TABLE entities ALTER COLUMN package DROP NOT NULL;
ALTER TABLE entities ALTER COLUMN package SET DEFAULT '';
ALTER TABLE entities ADD COLUMN IF NOT EXISTS doc_comment TEXT;
ALTER TABLE entities ADD COLUMN IF NOT EXISTS signature TEXT;
ALTER TABLE entities ADD COLUMN IF NOT EXISTS fields JSONB DEFAULT '[]'::jsonb;
ALTER TABLE entities ADD COLUMN IF NOT EXISTS methods JSONB DEFAULT '[]'::jsonb;
ALTER TABLE entities ADD COLUMN IF NOT EXISTS parameters JSONB DEFAULT '[]'::jsonb;
ALTER TABLE entities ADD COLUMN IF NOT EXISTS returns JSONB DEFAULT '[]'::jsonb;
ALTER TABLE entities ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}'::jsonb;
ALTER TABLE entities ADD COLUMN IF NOT EXISTS visibility TEXT DEFAULT 'public';
ALTER TABLE entities ADD COLUMN IF NOT EXISTS hash TEXT DEFAULT '';
ALTER TABLE entities ADD COLUMN IF NOT EXISTS is_exported BOOLEAN DEFAULT true;
ALTER TABLE entities ADD COLUMN IF NOT EXISTS is_test BOOLEAN DEFAULT false;
ALTER TABLE entities ADD COLUMN IF NOT EXISTS is_generated BOOLEAN DEFAULT false;
ALTER TABLE entities ADD COLUMN IF NOT EXISTS source_code TEXT DEFAULT '';
ALTER TABLE entities ADD COLUMN IF NOT EXISTS doc TEXT DEFAULT '';
ALTER TABLE entities ADD COLUMN IF NOT EXISTS comments TEXT DEFAULT '';
ALTER TABLE entities ADD COLUMN IF NOT EXISTS ast_hash TEXT DEFAULT '';
ALTER TABLE entities ADD COLUMN IF NOT EXISTS content_hash TEXT DEFAULT '';
ALTER TABLE entities ADD COLUMN IF NOT EXISTS type_params JSONB DEFAULT '[]'::jsonb;
ALTER TABLE entities ADD COLUMN IF NOT EXISTS implements JSONB DEFAULT '[]'::jsonb;
ALTER TABLE entities ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_entities_canonical_name ON entities (canonical_name);

-- 3. Relationships Schema Hardening
CREATE TABLE IF NOT EXISTS relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    analysis_id UUID,
    tenant_id UUID,
    workspace_id UUID,
    source_id UUID NOT NULL,
    target_id UUID NOT NULL,
    kind TEXT DEFAULT 'references' NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE relationships ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE relationships ADD COLUMN IF NOT EXISTS workspace_id UUID;

CREATE UNIQUE INDEX IF NOT EXISTS uq_relationships_edge 
  ON relationships (source_id, target_id, kind);
CREATE INDEX IF NOT EXISTS idx_relationships_src_target 
  ON relationships (source_id, target_id);
CREATE INDEX IF NOT EXISTS idx_relationships_tenant_ws 
  ON relationships (tenant_id, workspace_id);

-- 4. Cross-Repo Bridges Alignment (Using Active from_repo_id / to_repo_id)
DROP INDEX IF EXISTS idx_cross_repo_bridges;
CREATE INDEX IF NOT EXISTS idx_cross_repo_bridges 
  ON cross_repo_edges (from_repo_id, to_repo_id);