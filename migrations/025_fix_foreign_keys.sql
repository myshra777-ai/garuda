-- 025_fix_foreign_keys.sql
-- Ensure composite unique constraint exists on decisions(tenant_id, id)
ALTER TABLE decisions DROP CONSTRAINT IF EXISTS uq_decisions_tenant_id CASCADE;
ALTER TABLE decisions ADD CONSTRAINT uq_decisions_tenant_id UNIQUE (tenant_id, id);

-- Drop legacy foreign key and add composite key
ALTER TABLE decision_revisions
DROP CONSTRAINT IF EXISTS decision_revisions_decision_id_fkey,
DROP CONSTRAINT IF EXISTS fk_revisions_decisions;

ALTER TABLE decision_revisions
ADD CONSTRAINT decision_revisions_decision_id_fkey
FOREIGN KEY (tenant_id, decision_id)
REFERENCES decisions(tenant_id, id)
ON DELETE RESTRICT;
