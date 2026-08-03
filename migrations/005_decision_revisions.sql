-- Add decision revision history for audit and lineage comparison.
-- Note: decision_id matches decisions.id which is TEXT in the current schema.
CREATE TABLE IF NOT EXISTS decision_revisions (
    id UUID PRIMARY KEY,
    decision_id TEXT NOT NULL,  -- Changed from UUID to TEXT to match decisions.id
    revision_number INT NOT NULL,
    assumptions JSONB NOT NULL,
    facts JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_decision_revisions_decision ON decision_revisions(decision_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_decision_revisions_decision_revision ON decision_revisions(decision_id, revision_number);