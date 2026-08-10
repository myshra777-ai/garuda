-- Policy Engine for developer intent locking
CREATE TABLE IF NOT EXISTS policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    statement TEXT NOT NULL,
    scope_domain TEXT NOT NULL,
    scope_system TEXT NOT NULL,
    actor TEXT NOT NULL,                    -- Who set the policy
    status TEXT NOT NULL DEFAULT 'active', -- active, superseded, expired
    valid_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    valid_to TIMESTAMPTZ,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    superseded_by UUID REFERENCES policies(id) ON DELETE SET NULL,
    merkle_hash TEXT
);

CREATE INDEX idx_policies_tenant ON policies(tenant_id);
CREATE INDEX idx_policies_scope ON policies(tenant_id, scope_domain, scope_system);
CREATE INDEX idx_policies_status ON policies(tenant_id, status);
CREATE INDEX idx_policies_valid ON policies(tenant_id, valid_from, valid_to);

-- Policy violation log (audit trail)
CREATE TABLE IF NOT EXISTS policy_violations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    policy_id UUID REFERENCES policies(id) ON DELETE CASCADE,
    actor TEXT NOT NULL,
    attempted_action TEXT NOT NULL,
    decision_id UUID,                     -- The decision that was rejected
    reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_policy_violations_tenant ON policy_violations(tenant_id);
CREATE INDEX idx_policy_violations_policy ON policy_violations(policy_id);