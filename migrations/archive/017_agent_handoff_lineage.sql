-- ============================================================
-- Phase 5: Multi-Agent Handoff & Lineage DAG Tracking
-- ============================================================

-- 1. Agents Registry
CREATE TABLE IF NOT EXISTS agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name TEXT NOT NULL,
    model_type TEXT NOT NULL,
    session_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'idle', -- idle, working, transitioning, paused, offline
    current_task_id UUID,
    last_heartbeat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Tasks Registry
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'pending', -- pending, in_progress, paused, completed, abandoned
    priority INT NOT NULL DEFAULT 0,
    owner_agent_id UUID REFERENCES agents(id) ON DELETE SET NULL,
    parent_task_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
    scope_domain TEXT DEFAULT 'general',
    scope_system TEXT DEFAULT 'default',
    version INT NOT NULL DEFAULT 1, -- Optimistic locking
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

-- 3. Checkpoints (with CAS deduplication)
CREATE TABLE IF NOT EXISTS checkpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    agent_id UUID REFERENCES agents(id) ON DELETE SET NULL,
    state_hash TEXT NOT NULL, -- SHA-256 for CAS deduplication
    checkpoint_data JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'active', -- active, restored, expired
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    restored_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ
);

-- 4. Handoffs (with SAGA state machine)
CREATE TABLE IF NOT EXISTS handoffs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    source_agent_id UUID REFERENCES agents(id) ON DELETE RESTRICT,
    target_agent_id UUID REFERENCES agents(id) ON DELETE RESTRICT,
    checkpoint_id UUID REFERENCES checkpoints(id) ON DELETE RESTRICT,
    reason TEXT,
    status TEXT NOT NULL DEFAULT 'pending', -- pending, in_progress, completed, failed, rolled_back
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

-- 5. Lineage DAG (with cycle guard)
CREATE TABLE IF NOT EXISTS lineage_edges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    source_task_id UUID REFERENCES tasks(id) ON DELETE RESTRICT,
    target_task_id UUID REFERENCES tasks(id) ON DELETE RESTRICT,
    edge_type TEXT NOT NULL, -- 'handoff', 'depends_on', 'supersedes'
    handoff_id UUID REFERENCES handoffs(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT no_self_loop CHECK (source_task_id <> target_task_id)
);

-- 6. Performance Indexes
CREATE INDEX IF NOT EXISTS idx_agents_tenant_status ON agents(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_agents_session ON agents(session_id);
CREATE INDEX IF NOT EXISTS idx_tasks_tenant_owner ON tasks(tenant_id, owner_agent_id);
CREATE INDEX IF NOT EXISTS idx_tasks_tenant_status ON tasks(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_checkpoints_task_state ON checkpoints(task_id, state_hash);
CREATE INDEX IF NOT EXISTS idx_checkpoints_tenant ON checkpoints(tenant_id);
CREATE INDEX IF NOT EXISTS idx_handoffs_tenant_task ON handoffs(tenant_id, task_id);
CREATE INDEX IF NOT EXISTS idx_handoffs_status ON handoffs(status);
CREATE INDEX IF NOT EXISTS idx_lineage_graph ON lineage_edges(tenant_id, source_task_id, target_task_id);

-- 7. Cycle Prevention: Trigger to reject cycles before insertion
