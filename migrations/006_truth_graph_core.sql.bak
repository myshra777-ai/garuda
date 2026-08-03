-- Phase A0-1: Canonical truth graph core tables with tenant isolation.

-- 1. Evidence store (content-addressable)
CREATE TABLE evidence_store (
    tenant_id UUID NOT NULL,
    block_hash BYTEA NOT NULL,
    content TEXT NOT NULL,
    ref_count INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, block_hash)
);

-- 2. Facts
CREATE TABLE facts (
    tenant_id UUID NOT NULL,
    id UUID NOT NULL,
    statement TEXT NOT NULL,
    confidence FLOAT NOT NULL DEFAULT 0.8,
    evidence_ids BYTEA[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, id)
);

-- 3. Assumptions
CREATE TABLE assumptions (
    tenant_id UUID NOT NULL,
    id UUID NOT NULL,
    statement TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    confidence FLOAT NOT NULL DEFAULT 0.5,
    evidence_ids BYTEA[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, id)
);

-- 4. Decisions
CREATE TABLE decisions (
    tenant_id UUID NOT NULL,
    id UUID NOT NULL,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    fingerprint TEXT NOT NULL,
    evidence_ids BYTEA[],
    temporal_metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, id)
);

-- 5. Decision edges (lineage)
CREATE TABLE decision_edges (
    tenant_id UUID NOT NULL,
    from_decision_id UUID NOT NULL,
    to_decision_id UUID NOT NULL,
    edge_type TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, from_decision_id, to_decision_id)
);

-- 6. Decision revisions (audit)
CREATE TABLE decision_revisions (
    tenant_id UUID NOT NULL,
    id UUID NOT NULL,
    decision_id UUID NOT NULL,
    revision_number INT NOT NULL,
    snapshot_json JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, id)
);

-- 7. Contradictions
CREATE TABLE contradictions (
    tenant_id UUID NOT NULL,
    id UUID NOT NULL,
    node_a UUID NOT NULL,
    node_b UUID NOT NULL,
    authority_a TEXT NOT NULL,
    authority_b TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, id)
);

-- 8. Budgets (for phase A6.5)
CREATE TABLE budgets (
    tenant_id UUID PRIMARY KEY,
    remaining_tokens BIGINT NOT NULL,
    last_reset TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 9. Budget ledger (for audit)
CREATE TABLE budget_ledger (
    ledger_id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    tokens_consumed BIGINT NOT NULL,
    remaining_after BIGINT NOT NULL,
    consumed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    correlation_id UUID NOT NULL,
    decision_id UUID,
    error_code TEXT
);

-- Enable RLS on all tenant-owned tables
ALTER TABLE evidence_store ENABLE ROW LEVEL SECURITY;
ALTER TABLE facts ENABLE ROW LEVEL SECURITY;
ALTER TABLE assumptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE decisions ENABLE ROW LEVEL SECURITY;
ALTER TABLE decision_edges ENABLE ROW LEVEL SECURITY;
ALTER TABLE decision_revisions ENABLE ROW LEVEL SECURITY;
ALTER TABLE contradictions ENABLE ROW LEVEL SECURITY;
ALTER TABLE budgets ENABLE ROW LEVEL SECURITY;
ALTER TABLE budget_ledger ENABLE ROW LEVEL SECURITY;

-- RLS policies: allow access only when tenant_id matches current tenant
CREATE POLICY tenant_isolation_evidence_store ON evidence_store USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY tenant_isolation_facts ON facts USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY tenant_isolation_assumptions ON assumptions USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY tenant_isolation_decisions ON decisions USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY tenant_isolation_decision_edges ON decision_edges USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY tenant_isolation_decision_revisions ON decision_revisions USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY tenant_isolation_contradictions ON contradictions USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY tenant_isolation_budgets ON budgets USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY tenant_isolation_budget_ledger ON budget_ledger USING (tenant_id = current_setting('app.current_tenant')::UUID);

-- Force RLS for additional safety
ALTER TABLE evidence_store FORCE ROW LEVEL SECURITY;
ALTER TABLE facts FORCE ROW LEVEL SECURITY;
ALTER TABLE assumptions FORCE ROW LEVEL SECURITY;
ALTER TABLE decisions FORCE ROW LEVEL SECURITY;
ALTER TABLE decision_edges FORCE ROW LEVEL SECURITY;
ALTER TABLE decision_revisions FORCE ROW LEVEL SECURITY;
ALTER TABLE contradictions FORCE ROW LEVEL SECURITY;
ALTER TABLE budgets FORCE ROW LEVEL SECURITY;
ALTER TABLE budget_ledger FORCE ROW LEVEL SECURITY;

-- Indexes for performance
CREATE INDEX idx_decisions_tenant_status ON decisions(tenant_id, status);
CREATE INDEX idx_decisions_tenant_fingerprint ON decisions(tenant_id, fingerprint);
CREATE INDEX idx_decision_edges_to ON decision_edges(tenant_id, to_decision_id);
CREATE INDEX idx_contradictions_status ON contradictions(tenant_id, status);
