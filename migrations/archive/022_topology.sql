-- Phase 2: Topology Synthesis & Role Boundary Engine

CREATE TYPE agent_role_enum AS ENUM ('ARCHITECT', 'ENGINEER', 'AUDITOR', 'CUSTOM');
CREATE TYPE topology_status_enum AS ENUM ('PENDING', 'ACTIVE', 'PAUSED', 'COMPLETED', 'FAILED', 'VETOED');
CREATE TYPE task_status_enum AS ENUM ('PENDING', 'IN_PROGRESS', 'COMPLETED', 'QUARANTINED', 'FAILED');

CREATE TABLE IF NOT EXISTS topologies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    goal TEXT NOT NULL,
    scope_domain VARCHAR(64) NOT NULL,
    scope_system VARCHAR(64) NOT NULL,
    status topology_status_enum NOT NULL DEFAULT 'PENDING',
    max_token_budget BIGINT NOT NULL,
    tokens_consumed BIGINT NOT NULL DEFAULT 0,
    merkle_root VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS topology_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topology_id UUID NOT NULL REFERENCES topologies(id) ON DELETE CASCADE,
    sequence_no INT NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    required_role agent_role_enum NOT NULL,
    assigned_to UUID,
    scope VARCHAR(128) NOT NULL,
    status task_status_enum NOT NULL DEFAULT 'PENDING',
    depends_on JSONB DEFAULT '[]'::jsonb,
    token_budget BIGINT NOT NULL,
    tokens_used BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS topology_handoffs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topology_id UUID NOT NULL REFERENCES topologies(id) ON DELETE CASCADE,
    from_task_id UUID NOT NULL REFERENCES topology_tasks(id),
    to_task_id UUID NOT NULL REFERENCES topology_tasks(id),
    from_agent_id UUID NOT NULL,
    to_agent_id UUID NOT NULL,
    state_delta TEXT NOT NULL,
    reason TEXT NOT NULL,
    merkle_proof VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS topology_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topology_id UUID NOT NULL REFERENCES topologies(id) ON DELETE CASCADE,
    actor TEXT NOT NULL,
    action TEXT NOT NULL,
    reason TEXT NOT NULL,
    task_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_topologies_tenant_status ON topologies(tenant_id, status);
CREATE INDEX idx_tasks_topology_sequence ON topology_tasks(topology_id, sequence_no);
CREATE INDEX idx_handoffs_topology ON topology_handoffs(topology_id);