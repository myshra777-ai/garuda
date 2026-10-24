CREATE TABLE IF NOT EXISTS facts (
    id UUID PRIMARY KEY,
    statement TEXT NOT NULL,
    confidence FLOAT NOT NULL DEFAULT 0.8,
    evidence_ids UUID[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS assumptions (
    id UUID PRIMARY KEY,
    statement TEXT NOT NULL,
    confidence FLOAT NOT NULL DEFAULT 0.5,
    evidence_ids UUID[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);