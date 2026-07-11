-- Add decision revision history for audit and lineage comparison.
CREATE TABLE decision_revisions (
    id UUID PRIMARY KEY,
    decision_id UUID NOT NULL REFERENCES decisions(id) ON DELETE CASCADE,
    revision_number INT NOT NULL,
    assumptions JSONB NOT NULL,
    facts JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_decision_revisions_decision ON decision_revisions(decision_id);
CREATE UNIQUE INDEX idx_decision_revisions_decision_revision ON decision_revisions(decision_id, revision_number);
