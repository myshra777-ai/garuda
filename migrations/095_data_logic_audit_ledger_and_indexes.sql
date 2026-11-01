-- 096_policy_evaluations_decision_check.sql
ALTER TABLE policy_evaluations DROP CONSTRAINT IF EXISTS policy_evaluations_decision_check;
ALTER TABLE policy_evaluations ADD CONSTRAINT policy_evaluations_decision_check
    CHECK (decision IN ('ALLOW', 'WARN', 'REVIEW', 'BLOCK', 'INCONCLUSIVE'));

CREATE INDEX IF NOT EXISTS idx_entities_ws_kind_name ON entities (workspace_id, kind, name);
CREATE INDEX IF NOT EXISTS idx_claims_ws_from_pred ON claims (workspace_id, from_entity_id, claim_type);
CREATE INDEX IF NOT EXISTS idx_claims_ws_to_pred ON claims (workspace_id, to_entity_id, claim_type);
CREATE INDEX IF NOT EXISTS idx_doc_claims_ws_status ON document_claims (tenant_id, workspace_id, status);
CREATE INDEX IF NOT EXISTS idx_doc_claims_ws_pred ON document_claims (workspace_id, predicate);
