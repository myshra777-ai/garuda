-- Migration 014: Ensure decisions has a clean unique constraint on (tenant_id, id) or (id)

DO $$
BEGIN
    -- Drop existing PK if it lacks tenant_id
    IF EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'decisions_pkey'
    ) THEN
        ALTER TABLE decisions DROP CONSTRAINT decisions_pkey;
    END IF;

    -- Add primary key on tenant_id, id
    ALTER TABLE decisions ADD CONSTRAINT decisions_pkey PRIMARY KEY (tenant_id, id);
END $$;