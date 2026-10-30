-- Migration 095: Data & Logic Audit Ledger and Performance Indexes
CREATE TABLE IF NOT EXISTS data_logic_audit_ledger (
    audit_id VARCHAR(128) PRIMARY KEY,
    workspace_id UUID NOT NULL,
    metric_name TEXT NOT NULL,
    team_a_name TEXT NOT NULL,
    team_a_reported_cr NUMERIC(12,2) NOT NULL DEFAULT 0,
    team_b_name TEXT NOT NULL,
    team_b_reported_cr NUMERIC(12,2) NOT NULL DEFAULT 0,
    variance_cr NUMERIC(12,2) NOT NULL DEFAULT 0,
    reconciled_canonical_cr NUMERIC(12,2) NOT NULL DEFAULT 0,
    root_cause_ast TEXT NOT NULL DEFAULT '',
    reconciliation_sql TEXT NOT NULL DEFAULT '',
    status VARCHAR(64) NOT NULL DEFAULT 'UNRECONCILED_DISCREPANCY',
    approved_by TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_entities_ws_kind_name ON entities (workspace_id, kind, name);
CREATE INDEX IF NOT EXISTS idx_claims_ws_from_pred ON claims (workspace_id, from_entity_id, claim_type);
CREATE INDEX IF NOT EXISTS idx_claims_ws_to_pred ON claims (workspace_id, to_entity_id, claim_type);
CREATE INDEX IF NOT EXISTS idx_doc_claims_ws_status ON document_claims (tenant_id, workspace_id, status);
