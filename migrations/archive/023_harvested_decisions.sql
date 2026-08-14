-- 023_harvested_decisions.sql
CREATE TABLE IF NOT EXISTS harvested_decisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    source_type TEXT NOT NULL CHECK (source_type IN ('git_commit', 'adr', 'slack', 'github_pr', 'email')),
    source_id TEXT NOT NULL,
    source_url TEXT,
    raw_text TEXT NOT NULL,
    extracted_decision TEXT,
    confidence FLOAT DEFAULT 0.7 CHECK (confidence >= 0 AND confidence <= 1),
    human_validated BOOLEAN DEFAULT FALSE,
    decision_id UUID,  -- No foreign key constraint
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_harvested_source ON harvested_decisions(tenant_id, source_type, source_id);
CREATE INDEX idx_harvested_confidence ON harvested_decisions(tenant_id, confidence);
CREATE INDEX idx_harvested_validated ON harvested_decisions(tenant_id, human_validated);
CREATE INDEX idx_harvested_created ON harvested_decisions(tenant_id, created_at DESC);
CREATE INDEX idx_harvested_decision ON harvested_decisions(tenant_id, decision_id);

-- RLS (Row Level Security) for multi-tenant isolation
ALTER TABLE harvested_decisions ENABLE ROW LEVEL SECURITY;
CREATE POLICY harvest_tenant_isolation ON harvested_decisions
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);