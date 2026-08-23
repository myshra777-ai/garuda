-- migrations/055_runtime_evidence_schema.sql
-- Substrate for OpenTelemetry span normalization, entity correlation, and tri-state claim verification

CREATE TABLE IF NOT EXISTS runtime_observations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
    tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
    trace_id VARCHAR(64) NOT NULL DEFAULT '',
    span_id VARCHAR(32) NOT NULL DEFAULT '',
    parent_span_id VARCHAR(32) DEFAULT '',
    service_name VARCHAR(128) NOT NULL DEFAULT '',
    operation VARCHAR(256) NOT NULL DEFAULT '',
    entity_id UUID REFERENCES entities(id) ON DELETE SET NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    duration_ms DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    status_code VARCHAR(32) NOT NULL DEFAULT 'OK',
    source VARCHAR(64) NOT NULL DEFAULT 'opentelemetry',
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    attributes JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE runtime_observations 
    ADD COLUMN IF NOT EXISTS workspace_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
    ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
    ADD COLUMN IF NOT EXISTS trace_id VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS span_id VARCHAR(32) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS parent_span_id VARCHAR(32) DEFAULT '',
    ADD COLUMN IF NOT EXISTS service_name VARCHAR(128) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS operation VARCHAR(256) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS entity_id UUID REFERENCES entities(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS duration_ms DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    ADD COLUMN IF NOT EXISTS status_code VARCHAR(32) NOT NULL DEFAULT 'OK',
    ADD COLUMN IF NOT EXISTS source VARCHAR(64) NOT NULL DEFAULT 'opentelemetry',
    ADD COLUMN IF NOT EXISTS confidence DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    ADD COLUMN IF NOT EXISTS attributes JSONB NOT NULL DEFAULT '{}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_runtime_obs_workspace_tenant ON runtime_observations(workspace_id, tenant_id);
CREATE INDEX IF NOT EXISTS idx_runtime_obs_trace ON runtime_observations(trace_id, span_id);
CREATE INDEX IF NOT EXISTS idx_runtime_obs_entity ON runtime_observations(entity_id);

CREATE TABLE IF NOT EXISTS runtime_edges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
    source_entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    target_entity_id UUID REFERENCES entities(id) ON DELETE SET NULL,
    raw_target VARCHAR(512) NOT NULL DEFAULT '',
    invocation_count BIGINT NOT NULL DEFAULT 1,
    error_count BIGINT NOT NULL DEFAULT 0,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_trace_id VARCHAR(64) NOT NULL DEFAULT '',
    CONSTRAINT unq_runtime_edge UNIQUE (workspace_id, tenant_id, source_entity_id, raw_target)
);

CREATE INDEX IF NOT EXISTS idx_runtime_edges_lookup ON runtime_edges(workspace_id, tenant_id, source_entity_id);

CREATE TABLE IF NOT EXISTS claim_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
    claim_id UUID,
    source_entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    target_entity_id UUID REFERENCES entities(id) ON DELETE SET NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'UNVERIFIED',
    reason VARCHAR(64) NOT NULL DEFAULT 'NO_RUNTIME_OBSERVATION',
    static_edge_exists BOOLEAN NOT NULL DEFAULT FALSE,
    runtime_observed_count BIGINT NOT NULL DEFAULT 0,
    last_evaluated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    evidence_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    CONSTRAINT unq_claim_verification UNIQUE (workspace_id, tenant_id, source_entity_id, target_entity_id)
);

CREATE INDEX IF NOT EXISTS idx_claim_verifications_status ON claim_verifications(workspace_id, tenant_id, status);
