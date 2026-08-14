-- Migration 009: Agent Checkpoints with Composite Tenant Primary Keys & Unique Names
CREATE TABLE IF NOT EXISTS agent_checkpoints (
    id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    agent_id TEXT NOT NULL,
    checkpoint_name TEXT NOT NULL DEFAULT 'manual_checkpoint',
    task_id UUID,
    checkpoint_data JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    PRIMARY KEY (tenant_id, id)
);

-- Unique constraint for tenant + checkpoint_name upserts
CREATE UNIQUE INDEX IF NOT EXISTS idx_agent_checkpoints_tenant_name 
    ON agent_checkpoints (tenant_id, checkpoint_name);

-- Performance Indices
CREATE INDEX IF NOT EXISTS idx_checkpoints_tenant 
    ON agent_checkpoints (tenant_id);

CREATE INDEX IF NOT EXISTS idx_checkpoints_agent 
    ON agent_checkpoints (tenant_id, agent_id);

CREATE INDEX IF NOT EXISTS idx_checkpoints_task 
    ON agent_checkpoints (tenant_id, task_id) 
    WHERE task_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_checkpoints_status 
    ON agent_checkpoints (tenant_id, status);