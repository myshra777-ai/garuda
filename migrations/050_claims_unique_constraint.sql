-- Migration 050: Add unique constraint on claims for idempotent upserts

-- 1. Deduplicate any existing identical claims
DELETE FROM claims a
USING claims b
WHERE a.ctid < b.ctid
  AND a.workspace_id = b.workspace_id
  AND a.from_entity_id = b.from_entity_id
  AND a.to_entity_id = b.to_entity_id
  AND a.claim_type = b.claim_type;

-- 2. Add composite unique constraint
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uq_claims_edge'
    ) THEN
        ALTER TABLE claims ADD CONSTRAINT uq_claims_edge 
        UNIQUE (workspace_id, from_entity_id, to_entity_id, claim_type);
    END IF;
EXCEPTION
    WHEN duplicate_table OR duplicate_object THEN NULL;
END $$;

-- 3. Ensure lookup index exists
CREATE INDEX IF NOT EXISTS idx_claims_workspace_lookup 
ON claims (workspace_id, from_entity_id, to_entity_id);