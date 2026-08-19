-- 033_contracts_and_impact.sql
-- API contracts and downstream impact radius schema

CREATE TABLE IF NOT EXISTS api_contracts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    workspace_id UUID,
    repository_id UUID REFERENCES repositories(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL,
    method TEXT NOT NULL,
    request_schema JSONB,
    response_schema JSONB,
    version TEXT NOT NULL DEFAULT 'v1',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_contracts_tenant ON api_contracts (tenant_id);
CREATE INDEX IF NOT EXISTS idx_api_contracts_workspace ON api_contracts (workspace_id);
CREATE INDEX IF NOT EXISTS idx_api_contracts_repo ON api_contracts (repository_id);
CREATE INDEX IF NOT EXISTS idx_api_contracts_endpoint ON api_contracts (tenant_id, endpoint, method);

CREATE TABLE IF NOT EXISTS impact_assessments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    workspace_id UUID,
    entity_id UUID NOT NULL,
    radius INT NOT NULL DEFAULT 1,
    impacted_nodes JSONB NOT NULL DEFAULT '[]',
    risk_score FLOAT NOT NULL DEFAULT 0.0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_impact_assessments_tenant ON impact_assessments (tenant_id);
CREATE INDEX IF NOT EXISTS idx_impact_assessments_entity ON impact_assessments (entity_id);
CREATE INDEX IF NOT EXISTS idx_impact_assessments_created ON impact_assessments (created_at DESC);

CREATE OR REPLACE FUNCTION update_api_contracts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_api_contracts_updated_at ON api_contracts;
CREATE TRIGGER trigger_api_contracts_updated_at
    BEFORE UPDATE ON api_contracts
    FOR EACH ROW
    EXECUTE FUNCTION update_api_contracts_updated_at();