CREATE TABLE IF NOT EXISTS audit_events (
    id SERIAL PRIMARY KEY,
    decision_id TEXT NOT NULL REFERENCES decisions(id) ON DELETE CASCADE,
    actor TEXT NOT NULL,
    old_status TEXT,
    new_status TEXT NOT NULL,
    reason TEXT,
    timestamp TIMESTAMPTZ NOT NULL,
    signature BYTEA
);

CREATE INDEX IF NOT EXISTS idx_audit_decision ON audit_events(decision_id);
CREATE INDEX IF NOT EXISTS idx_audit_actor ON audit_events(actor);
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_events(timestamp);