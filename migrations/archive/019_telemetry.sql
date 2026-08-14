-- Telemetry events table (anonymised, no PII)
CREATE TABLE IF NOT EXISTS telemetry_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_hash TEXT NOT NULL,
    session_id TEXT NOT NULL,
    mode TEXT NOT NULL, -- 'active' or 'passive'
    garuda_version TEXT,
    agent_runtime TEXT,

    -- Decisions
    decision_status TEXT,
    decision_scope JSONB,
    decision_confidence FLOAT,
    contradiction_resolved BOOLEAN,

    -- Model
    model_provider TEXT,
    model_name TEXT,
    model_route TEXT,

    -- Cost & Savings
    tokens_estimated BIGINT,
    tokens_saved BIGINT,
    cost_saved_usd FLOAT,
    budget_remaining BIGINT,

    -- Performance
    cold_start_latency_ms FLOAT,
    warm_start_latency_ms FLOAT,
    handoff_latency_ms FLOAT,
    verification_latency_ms FLOAT,

    -- Usage
    active_agents INT,
    total_handoffs BIGINT,
    total_contradictions BIGINT,
    total_decisions BIGINT,
    total_verifications BIGINT,
    budget_exhausted BOOLEAN,

    -- Coordination
    handoff_success_rate FLOAT,
    contradiction_reduction_rate FLOAT,
    token_reuse_rate FLOAT,
    duplicate_work_reduction FLOAT,
    coordination_score FLOAT,

    -- Hallucinations
    hallucinations_prevented BIGINT,
    hallucination_reduction_per_model FLOAT,

    -- Timestamp
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_telemetry_created ON telemetry_events(created_at DESC);
CREATE INDEX idx_telemetry_model ON telemetry_events(model_name);
CREATE INDEX idx_telemetry_mode ON telemetry_events(mode);
CREATE INDEX idx_telemetry_instance ON telemetry_events(instance_hash);