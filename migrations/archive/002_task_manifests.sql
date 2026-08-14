CREATE TYPE task_status AS ENUM ('not_started', 'in_progress', 'paused', 'completed', 'stale');

CREATE TABLE task_manifests (
    task_id UUID PRIMARY KEY,
    customer_id UUID NOT NULL,
    credential_ref TEXT NOT NULL,
    title TEXT NOT NULL,
    scope_domain TEXT NOT NULL,
    scope_system TEXT NOT NULL,
    status task_status NOT NULL DEFAULT 'not_started',
    manifest_blocks BYTEA[] NOT NULL,
    normalized_ir JSONB NOT NULL,
    ir_version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_manifest_lookup ON task_manifests(customer_id, scope_domain, scope_system);
CREATE INDEX idx_manifest_status ON task_manifests(status);
CREATE INDEX idx_manifest_updated ON task_manifests(updated_at);

-- Add expires_at column to task_manifests
ALTER TABLE task_manifests ADD COLUMN expires_at TIMESTAMPTZ;

-- Optional: index for faster expiry queries
CREATE INDEX idx_manifest_expires ON task_manifests(expires_at);