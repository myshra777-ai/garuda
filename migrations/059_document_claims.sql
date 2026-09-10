-- 059_document_claims.sql
-- Intent and documentation claims layer for Knowledge Drift detection

CREATE TABLE IF NOT EXISTS document_claims (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace VARCHAR(255) NOT NULL DEFAULT 'default',
    tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
    document_path TEXT NOT NULL,
    document_title TEXT NOT NULL,
    document_type VARCHAR(64) NOT NULL DEFAULT 'ADR', -- ADR, RFC, OPENAPI, README
    section_title TEXT,
    line_start INT NOT NULL,
    line_end INT NOT NULL,
    raw_statement TEXT NOT NULL,
    
    -- Normalized Claim IR Fields
    subject VARCHAR(255) NOT NULL,
    modality VARCHAR(32) NOT NULL, -- MUST, MUST_NOT, SHOULD, SHOULD_NOT, MAY
    predicate VARCHAR(64) NOT NULL, -- CALLS, IMPLEMENTS, ENFORCES, EXPOSES, IDEMPOTENT
    object VARCHAR(255) NOT NULL,
    
    -- Verification & Graph Linking
    provenance_class VARCHAR(32) NOT NULL DEFAULT 'EXTRACTED', -- EXTRACTED, DECLARED, INFERRED
    confidence NUMERIC(3, 2) NOT NULL DEFAULT 1.0,
    status VARCHAR(32) NOT NULL DEFAULT 'UNVERIFIED', -- SUPPORTED, UNVERIFIED, CONTRADICTED
    matched_entity_id UUID REFERENCES entities(id) ON DELETE SET NULL,
    contradiction_reason TEXT,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_doc_claims_lookup 
ON document_claims(tenant_id, workspace, status);

CREATE INDEX IF NOT EXISTS idx_doc_claims_subject 
ON document_claims(tenant_id, subject);

CREATE INDEX IF NOT EXISTS idx_doc_claims_path 
ON document_claims(document_path);
