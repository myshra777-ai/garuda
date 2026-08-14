-- Migration 010: Real-Time Token and Execution Budget Metering

CREATE TABLE IF NOT EXISTS tenant_budgets (
    tenant_id UUID PRIMARY KEY,
    token_balance BIGINT NOT NULL DEFAULT 1000000,
    tokens_consumed BIGINT NOT NULL DEFAULT 0,
    execution_limit INT NOT NULL DEFAULT 10000,
    executions_consumed INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'active',
    monthly_limit BIGINT NOT NULL DEFAULT 1000000,
    last_reset_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS budget_ledger (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    agent_id TEXT NOT NULL,
    task_id UUID,
    tokens_used INT NOT NULL DEFAULT 0,
    executions_used INT NOT NULL DEFAULT 1,
    operation TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for audit trailing and performance
CREATE INDEX IF NOT EXISTS idx_budget_ledger_tenant ON budget_ledger(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_budget_ledger_agent ON budget_ledger(tenant_id, agent_id);