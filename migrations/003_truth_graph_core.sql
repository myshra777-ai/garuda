-- Phase 0.5: Truth Graph Core Tables
-- This migration establishes the physical storage layout for the v4.0 architecture.

-- 1. Evidence Store (content-addressable)
CREATE TABLE IF NOT EXISTS evidence_store (
    block_hash BYTEA PRIMARY KEY,
    content TEXT NOT NULL,
    ref_count INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_evidence_created ON evidence_store(created_at);

-- 2. Facts (claims believed to be true)
CREATE TABLE IF NOT EXISTS facts (
    id UUID PRIMARY KEY,
    statement TEXT NOT NULL,
    confidence FLOAT NOT NULL DEFAULT 0.8,
    evidence_ids UUID[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Assumptions (claims that may be false)
CREATE TABLE IF NOT EXISTS assumptions (
    id UUID PRIMARY KEY,
    statement TEXT NOT NULL,
    confidence FLOAT NOT NULL DEFAULT 0.5,
    evidence_ids UUID[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. Decisions (the core unit of organizational cognition)
CREATE TABLE IF NOT EXISTS decisions (
    id UUID PRIMARY KEY,
    customer_id UUID NOT NULL,
    statement TEXT NOT NULL,
    assumptions JSONB NOT NULL,
    facts JSONB NOT NULL,
    evidence_ids BYTEA[] NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('draft', 'approved', 'executed', 'stale')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    parent_id UUID REFERENCES decisions(id) ON DELETE SET NULL,
    child_ids UUID[] DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_decisions_customer ON decisions(customer_id);
CREATE INDEX IF NOT EXISTS idx_decisions_status ON decisions(status);
CREATE INDEX IF NOT EXISTS idx_decisions_parent ON decisions(parent_id);
CREATE INDEX IF NOT EXISTS idx_decisions_created ON decisions(created_at);

-- 5. Contradictions (first-class conflict objects)
CREATE TABLE IF NOT EXISTS contradictions (
    id UUID PRIMARY KEY,
    fact_a_id UUID REFERENCES facts(id) ON DELETE CASCADE,
    fact_b_id UUID REFERENCES facts(id) ON DELETE CASCADE,
    severity TEXT NOT NULL CHECK (severity IN ('low', 'medium', 'high')),
    resolved BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_contradictions_fact_a ON contradictions(fact_a_id);
CREATE INDEX IF NOT EXISTS idx_contradictions_fact_b ON contradictions(fact_b_id);
CREATE INDEX IF NOT EXISTS idx_contradictions_resolved ON contradictions(resolved);