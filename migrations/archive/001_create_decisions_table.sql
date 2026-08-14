CREATE TABLE IF NOT EXISTS decisions (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL,
    parent_id TEXT REFERENCES decisions(id) ON DELETE SET NULL,
    decision TEXT NOT NULL,
    rationale TEXT,
    alternatives JSONB,
    scope JSONB NOT NULL,
    status TEXT NOT NULL,
    confidence FLOAT,
    owner TEXT NOT NULL,
    approvers JSONB,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ,
    approved_at TIMESTAMPTZ,
    dependencies JSONB,
    affected_systems JSONB,
    authorized_by TEXT,
    signature BYTEA,
    CONSTRAINT valid_status CHECK (status IN ('DRAFT', 'REVIEW', 'APPROVED', 'CANONICAL', 'SUPERSEDED', 'ARCHIVED'))
);

CREATE INDEX IF NOT EXISTS idx_decisions_scope ON decisions USING GIN (scope);
CREATE INDEX IF NOT EXISTS idx_decisions_status ON decisions(status);
CREATE INDEX IF NOT EXISTS idx_decisions_owner ON decisions(owner);
CREATE INDEX IF NOT EXISTS idx_decisions_parent ON decisions(parent_id);