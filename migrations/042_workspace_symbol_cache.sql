-- Migration 042: Workspace Multi-Module Ingestion and Symbol Cache

-- 1. Workspace Registry
CREATE TABLE IF NOT EXISTS workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name TEXT NOT NULL,
    root_path TEXT NOT NULL,
    is_go_work BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_workspace_tenant_name UNIQUE (tenant_id, name)
);

CREATE INDEX IF NOT EXISTS idx_workspaces_tenant ON workspaces(tenant_id);

-- 2. Workspace Modules (tracks tree hashes for incremental cache invalidation)
CREATE TABLE IF NOT EXISTS workspace_modules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    module_path TEXT NOT NULL,
    relative_path TEXT NOT NULL,
    commit_sha TEXT NOT NULL DEFAULT '',
    tree_hash BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_workspace_module UNIQUE (workspace_id, module_path)
);

CREATE INDEX IF NOT EXISTS idx_workspace_modules_lookup ON workspace_modules(tenant_id, module_path);

-- 3. Content-Addressed Symbol Cache (extracted entities & AST signatures)
CREATE TABLE IF NOT EXISTS symbol_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    package_path TEXT NOT NULL,
    symbol_name TEXT NOT NULL,
    kind TEXT NOT NULL,
    receiver TEXT NOT NULL DEFAULT '',
    signature_hash BYTEA NOT NULL,
    ast_hash BYTEA NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_symbol_entry UNIQUE (tenant_id, package_path, receiver, symbol_name)
);

CREATE INDEX IF NOT EXISTS idx_symbol_cache_lookup ON symbol_cache(tenant_id, package_path);
CREATE INDEX IF NOT EXISTS idx_symbol_cache_kind ON symbol_cache(tenant_id, kind);

-- 4. Cross-Module Dependency Edges
CREATE TABLE IF NOT EXISTS cross_module_edges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    from_module TEXT NOT NULL,
    from_package TEXT NOT NULL,
    from_symbol TEXT NOT NULL,
    to_module TEXT NOT NULL,
    to_package TEXT NOT NULL,
    to_symbol TEXT NOT NULL,
    edge_type TEXT NOT NULL,
    confidence FLOAT NOT NULL DEFAULT 1.0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_cross_module_edge UNIQUE (workspace_id, from_package, from_symbol, to_package, to_symbol, edge_type)
);

CREATE INDEX IF NOT EXISTS idx_cross_module_from ON cross_module_edges(tenant_id, from_package, from_symbol);
CREATE INDEX IF NOT EXISTS idx_cross_module_to ON cross_module_edges(tenant_id, to_package, to_symbol);