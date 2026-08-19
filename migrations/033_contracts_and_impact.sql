-- 033_contracts_and_impact.sql
-- Cross-repo contract detection and impact analysis

-- 1. API Contracts table – stores extracted API contracts
CREATE TABLE IF NOT EXISTS api_contracts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    workspace_id UUID NOT NULL,
    repository_id UUID NOT NULL,
    analysis_id UUID NOT NULL REFERENCES decision_revisions(id),
    service_name TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    method TEXT NOT NULL, -- GET, POST, PUT, DELETE, etc.
    request_schema JSONB,
    response_schema JSONB,
    contract_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, workspace_id, repository_id, service_name, endpoint, method)
);

CREATE INDEX idx_api_contracts_tenant ON api_contracts(tenant_id);
CREATE INDEX idx_api_contracts_workspace ON api_contracts(workspace_id);
CREATE INDEX idx_api_contracts_service ON api_contracts(service_name);
CREATE INDEX idx_api_contracts_endpoint ON api_contracts(endpoint);

-- 2. Contract Consumers – who calls which API
CREATE TABLE IF NOT EXISTS contract_consumers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    workspace_id UUID NOT NULL,
    contract_id UUID NOT NULL REFERENCES api_contracts(id) ON DELETE CASCADE,
    consumer_entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    consumer_repo_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    evidence JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, contract_id, consumer_entity_id)
);

CREATE INDEX idx_contract_consumers_contract ON contract_consumers(contract_id);
CREATE INDEX idx_contract_consumers_consumer ON contract_consumers(consumer_entity_id);

-- 3. Impact Analysis Reports – store historical impact analyses
CREATE TABLE IF NOT EXISTS impact_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    workspace_id UUID NOT NULL,
    analysis_id UUID NOT NULL REFERENCES decision_revisions(id),
    baseline_snapshot JSONB NOT NULL,
    proposed_snapshot JSONB NOT NULL,
    breaking_changes JSONB NOT NULL,
    warnings JSONB NOT NULL,
    evidence_root TEXT NOT NULL, -- Merkle root of evidence
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_impact_reports_tenant ON impact_reports(tenant_id);
CREATE INDEX idx_impact_reports_workspace ON impact_reports(workspace_id);