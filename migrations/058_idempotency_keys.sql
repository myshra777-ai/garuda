-- Idempotency Keys (Multi-tenant)
CREATE TABLE IF NOT EXISTS idempotency_keys (
    tenant_id UUID NOT NULL,
    idempotency_key TEXT NOT NULL,
    decision_id UUID NOT NULL,
    revision_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_idempotency_keys_tenant 
ON idempotency_keys (tenant_id, idempotency_key);
