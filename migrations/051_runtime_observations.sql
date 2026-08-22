-- migrations/051_runtime_observations.sql
-- Lightweight runtime trace observation substrate for Garuda Claims

CREATE TABLE IF NOT EXISTS runtime_observations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    trace_id VARCHAR(64) NOT NULL,
    span_id VARCHAR(32) NOT NULL,
    source_service VARCHAR(128) NOT NULL,
    target_service VARCHAR(128) NOT NULL,
    source_entity_id UUID REFERENCES entities(id) ON DELETE SET NULL,
    target_entity_id UUID REFERENCES entities(id) ON DELETE SET NULL,
    operation VARCHAR(255) NOT NULL,
    status_code VARCHAR(16) NOT NULL DEFAULT 'OK',
    duration_ms INT NOT NULL DEFAULT 0,
    environment VARCHAR(32) NOT NULL DEFAULT 'production',
    observed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_runtime_obs_ws ON runtime_observations(workspace_id, observed_at DESC);
CREATE INDEX IF NOT EXISTS idx_runtime_obs_services ON runtime_observations(source_service, target_service);