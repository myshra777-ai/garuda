-- migrations/054_add_tenant_id_to_revisions.sql
-- Ensure tenant_id exists across decision revision tables

ALTER TABLE IF EXISTS decision_revisions 
ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001';

CREATE INDEX IF NOT EXISTS idx_decision_revisions_tenant ON decision_revisions(tenant_id);
