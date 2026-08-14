-- Add tenant_id column if missing
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'repositories' AND column_name = 'tenant_id'
    ) THEN
        ALTER TABLE repositories ADD COLUMN tenant_id UUID NOT NULL DEFAULT gen_random_uuid();
        -- If you want to enforce tenant isolation, add a foreign key later:
        -- ALTER TABLE repositories ADD CONSTRAINT fk_repositories_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id);
    END IF;
END $$;