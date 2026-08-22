-- Migration 044: Align decisions table columns with analysis store requirements

ALTER TABLE decisions ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS title TEXT;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS statement TEXT;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS rationale TEXT;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS status TEXT DEFAULT 'PROPOSED';
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS scope_domain TEXT DEFAULT 'general';
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS scope_system TEXT DEFAULT 'cli';
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS scope JSONB DEFAULT '{}'::jsonb;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS owner TEXT DEFAULT 'system';
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS confidence DOUBLE PRECISION DEFAULT 1.0;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS fingerprint TEXT;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS valid_from TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();

-- Required for ON CONFLICT (tenant_id, id) resolution
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uq_decisions_tenant_id'
    ) THEN
        ALTER TABLE decisions ADD CONSTRAINT uq_decisions_tenant_id UNIQUE (tenant_id, id);
    END IF;
EXCEPTION
    WHEN duplicate_table OR duplicate_object THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_decisions_tenant_lookup ON decisions (tenant_id, scope_domain, scope_system);