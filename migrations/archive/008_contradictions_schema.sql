-- 008_contradictions_schema.sql
-- Drop and recreate contradictions table with correct foreign keys
DROP TABLE IF EXISTS contradictions CASCADE;

CREATE TABLE contradictions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    decision_a UUID NOT NULL,
    decision_b UUID NOT NULL,
    severity TEXT NOT NULL,
    resolved BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    -- Composite foreign keys: tenant_id + decision_a/b references decisions(tenant_id, id)
    CONSTRAINT fk_contradictions_decision_a 
        FOREIGN KEY (tenant_id, decision_a) REFERENCES decisions(tenant_id, id) ON DELETE CASCADE,
    CONSTRAINT fk_contradictions_decision_b 
        FOREIGN KEY (tenant_id, decision_b) REFERENCES decisions(tenant_id, id) ON DELETE CASCADE
);

CREATE INDEX idx_contradictions_tenant ON contradictions(tenant_id);
CREATE INDEX idx_contradictions_resolved ON contradictions(tenant_id, resolved);
