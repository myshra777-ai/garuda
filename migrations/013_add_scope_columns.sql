-- Migration 013: Ensure scope_domain and scope_system exist on decisions table

ALTER TABLE decisions ADD COLUMN IF NOT EXISTS scope_domain TEXT DEFAULT '';
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS scope_system TEXT DEFAULT '';

-- Create index for scope filtering queries
CREATE INDEX IF NOT EXISTS idx_decisions_scope ON decisions(tenant_id, scope_domain, scope_system);