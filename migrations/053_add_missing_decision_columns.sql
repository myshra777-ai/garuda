-- migrations/053_add_missing_decision_columns.sql
-- Add missing scope columns to decisions table

ALTER TABLE decisions 
ADD COLUMN IF NOT EXISTS system VARCHAR(128) NOT NULL DEFAULT 'default',
ADD COLUMN IF NOT EXISTS domain VARCHAR(128) NOT NULL DEFAULT 'architecture',
ADD COLUMN IF NOT EXISTS service VARCHAR(128) NOT NULL DEFAULT 'default',
ADD COLUMN IF NOT EXISTS component VARCHAR(128) NOT NULL DEFAULT 'default',
ADD COLUMN IF NOT EXISTS layer VARCHAR(64) NOT NULL DEFAULT 'backend';

CREATE INDEX IF NOT EXISTS idx_decisions_system_domain ON decisions(tenant_id, system, domain);
