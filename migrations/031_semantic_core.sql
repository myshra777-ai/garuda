-- 031_semantic_core.sql
-- Semantic Graph – Entities, Claims, Observations
-- This builds the Company Brain foundation

-- 1. Entities table – every discovered type, struct, interface, function
CREATE TABLE IF NOT EXISTS entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    workspace_id UUID NOT NULL,
    repository_id UUID NOT NULL,
    analysis_id UUID NOT NULL, -- references decision_revisions.id
    name TEXT NOT NULL,
    kind TEXT NOT NULL, -- 'struct', 'interface', 'function', 'type', 'api', 'service'
    package TEXT NOT NULL,
    file_path TEXT NOT NULL,
    commit_sha TEXT NOT NULL,
    signature TEXT, -- For functions: the full signature
    fields JSONB DEFAULT '[]'::jsonb, -- For structs: list of fields
    methods JSONB DEFAULT '[]'::jsonb, -- For interfaces/structs: list of methods
    is_exported BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, workspace_id, repository_id, name, package)
);

CREATE INDEX idx_entities_tenant ON entities(tenant_id);
CREATE INDEX idx_entities_workspace ON entities(workspace_id);
CREATE INDEX idx_entities_repository ON entities(repository_id);
CREATE INDEX idx_entities_name ON entities(name);
CREATE INDEX idx_entities_kind ON entities(kind);
CREATE INDEX idx_entities_package ON entities(package);

-- 2. Claims table – relationships between entities
CREATE TABLE IF NOT EXISTS claims (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    workspace_id UUID NOT NULL,
    repository_id UUID NOT NULL,
    analysis_id UUID NOT NULL, -- references decision_revisions.id
    from_entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    to_entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    claim_type TEXT NOT NULL, -- 'imports', 'calls', 'depends_on', 'implements', 'used_by', 'api_consumer'
    details JSONB DEFAULT '{}'::jsonb, -- Additional metadata (e.g., call count, usage context)
    confidence FLOAT DEFAULT 0.9,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, from_entity_id, to_entity_id, claim_type)
);

CREATE INDEX idx_claims_tenant ON claims(tenant_id);
CREATE INDEX idx_claims_from ON claims(from_entity_id);
CREATE INDEX idx_claims_to ON claims(to_entity_id);
CREATE INDEX idx_claims_type ON claims(claim_type);
CREATE INDEX idx_claims_analysis ON claims(analysis_id);

-- 3. Observations – evidence for each claim
CREATE TABLE IF NOT EXISTS observations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    claim_id UUID NOT NULL REFERENCES claims(id) ON DELETE CASCADE,
    evidence_hash BYTEA NOT NULL, -- references evidence_store.block_hash
    evidence_text TEXT, -- Optional snippet
    observed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, claim_id, evidence_hash)
);

CREATE INDEX idx_observations_claim ON observations(claim_id);
CREATE INDEX idx_observations_hash ON observations(evidence_hash);
CREATE INDEX idx_observations_tenant ON observations(tenant_id);

-- 4. RLS Policies (optional – enable if you use PostgreSQL RLS)
ALTER TABLE entities ENABLE ROW LEVEL SECURITY;
ALTER TABLE claims ENABLE ROW LEVEL SECURITY;
ALTER TABLE observations ENABLE ROW LEVEL SECURITY;

-- Note: For RLS to work, you need to set app.current_tenant_id in each session
-- For now, we'll use application-level tenant isolation