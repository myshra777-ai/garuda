CREATE TABLE contradictions (
    id UUID PRIMARY KEY,
    fact_a_id UUID REFERENCES facts(id),
    fact_b_id UUID REFERENCES facts(id),
    severity TEXT NOT NULL, -- "low", "medium", "high"
    resolved BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);