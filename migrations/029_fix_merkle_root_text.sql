-- Idempotent migration: convert BYTEA to TEXT (hex) if needed
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'merkle_roots'
          AND column_name = 'root_hash'
          AND data_type = 'bytea'
    ) THEN
        ALTER TABLE merkle_roots
            ALTER COLUMN root_hash TYPE TEXT USING encode(root_hash, 'hex');
    END IF;
END $$;