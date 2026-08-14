-- migrations/028_semantic_core.sql

-- 1. Extend workspaces with semantic metadata
ALTER TABLE workspaces ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE workspaces ADD COLUMN IF NOT EXISTS created_by TEXT;

-- 2. Extend repositories with analysis tracking
ALTER TABLE repositories ADD COLUMN IF NOT EXISTS language TEXT;
ALTER TABLE repositories ADD COLUMN IF NOT EXISTS current_commit TEXT;
ALTER TABLE repositories ADD COLUMN IF NOT EXISTS last_analyzed_at TIMESTAMPTZ;
ALTER TABLE repositories ADD COLUMN IF NOT EXISTS analysis_status TEXT DEFAULT 'pending';

-- 3. Analysis artifacts (immutable snapshots)
CREATE TABLE IF NOT EXISTS analysis_artifacts (
    id UUID PRIMARY KEY,
    repository_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    commit_sha TEXT NOT NULL,
    analyzer_name TEXT NOT NULL,
    analyzer_version TEXT NOT NULL,
    source_fingerprint TEXT NOT NULL,
    semantic_fingerprint TEXT NOT NULL,
    schema_version INT NOT NULL,
    status TEXT NOT NULL, -- 'SUCCESS', 'PARTIAL', 'FAILED'
    started_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    model JSONB NOT NULL,
    error_summary TEXT,
    created_at TIMESTAMPTZ NOT NULL
);

-- 4. Entities (projected for fast queries)
CREATE TABLE IF NOT EXISTS entities (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    repository_id UUID NOT NULL,
    canonical_urn TEXT NOT NULL,
    name TEXT NOT NULL,
    kind TEXT NOT NULL, -- 'STRUCT', 'INTERFACE', 'SERVICE', 'API', 'SCHEMA', etc.
    package TEXT,
    first_seen_commit TEXT NOT NULL,
    last_seen_commit TEXT NOT NULL,
    latest_snapshot_id UUID REFERENCES analysis_artifacts(id),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE(tenant_id, canonical_urn)
);

-- 5. Claims (projected semantic statements)
CREATE TABLE IF NOT EXISTS claims (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    subject_entity_id UUID REFERENCES entities(id) ON DELETE CASCADE,
    predicate TEXT NOT NULL,
    object_entity_id UUID REFERENCES entities(id) ON DELETE CASCADE,
    class TEXT NOT NULL, -- 'OBSERVED_CODE', 'OBSERVED_RUNTIME', 'INFERRED', 'DECISION', 'POLICY', 'VERIFIED'
    confidence FLOAT,
    status TEXT NOT NULL, -- 'PROPOSED', 'VERIFIED', 'CONTRADICTED', 'DEPRECATED'
    evidence_ids UUID[],
    valid_from TIMESTAMPTZ NOT NULL,
    valid_to TIMESTAMPTZ,
    snapshot_id UUID REFERENCES analysis_artifacts(id),
    created_at TIMESTAMPTZ NOT NULL
);

-- Indexes
CREATE INDEX idx_artifacts_repo_commit ON analysis_artifacts(repository_id, commit_sha);
CREATE INDEX idx_artifacts_status ON analysis_artifacts(status);
CREATE INDEX idx_entities_tenant_repo ON entities(tenant_id, repository_id);
CREATE INDEX idx_entities_urn ON entities(canonical_urn);
CREATE INDEX idx_claims_tenant_subject ON claims(tenant_id, subject_entity_id);
CREATE INDEX idx_claims_tenant_object ON claims(tenant_id, object_entity_id);
CREATE INDEX idx_claims_snapshot ON claims(snapshot_id);