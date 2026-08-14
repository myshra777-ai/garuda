-- 027_hardening.sql (Idempotent)
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS workspace_id UUID;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS repository_id UUID;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS commit_sha TEXT;

ALTER TABLE evidence_store ADD COLUMN IF NOT EXISTS tenant_id UUID;

-- Add composite primary key to evidence_store
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints tc
        JOIN information_schema.constraint_column_usage ccu
        ON tc.constraint_name = ccu.constraint_name
        WHERE tc.table_name = 'evidence_store'
          AND tc.constraint_type = 'PRIMARY KEY'
          AND ccu.column_name = 'block_hash'
          AND NOT EXISTS (
              SELECT 1 FROM information_schema.constraint_column_usage
              WHERE constraint_name = tc.constraint_name
                AND column_name = 'tenant_id'
          )
    ) THEN
        ALTER TABLE evidence_store DROP CONSTRAINT evidence_store_pkey;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_name = 'evidence_store'
          AND constraint_type = 'PRIMARY KEY'
    ) THEN
        ALTER TABLE evidence_store ADD PRIMARY KEY (tenant_id, block_hash);
    END IF;
END $$;
