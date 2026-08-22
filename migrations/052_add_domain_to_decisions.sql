-- migrations/052_add_domain_to_decisions.sql
-- Ensure domain column exists on decisions table for scope classification

ALTER TABLE decisions 
ADD COLUMN IF NOT EXISTS domain VARCHAR(128) NOT NULL DEFAULT 'architecture';

CREATE INDEX IF NOT EXISTS idx_decisions_domain ON decisions(tenant_id, domain);
