-- Migration 045: Safely relax legacy constraints and align decision_revisions schema

-- 1. Safely drop NOT NULL on legacy columns if they exist in decisions
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'decisions' AND column_name = 'domain') THEN
        ALTER TABLE decisions ALTER COLUMN domain DROP NOT NULL;
        ALTER TABLE decisions ALTER COLUMN domain SET DEFAULT 'architecture';
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'decisions' AND column_name = 'system') THEN
        ALTER TABLE decisions ALTER COLUMN system DROP NOT NULL;
        ALTER TABLE decisions ALTER COLUMN system SET DEFAULT 'ast-analyzer';
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'decisions' AND column_name = 'team') THEN
        ALTER TABLE decisions ALTER COLUMN team DROP NOT NULL;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'decisions' AND column_name = 'env') THEN
        ALTER TABLE decisions ALTER COLUMN env DROP NOT NULL;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'decisions' AND column_name = 'decision') THEN
        ALTER TABLE decisions ALTER COLUMN decision DROP NOT NULL;
    END IF;
END $$;

-- 2. Align decision_revisions table columns
ALTER TABLE decision_revisions ADD COLUMN IF NOT EXISTS facts JSONB DEFAULT '[]'::jsonb;
ALTER TABLE decision_revisions ADD COLUMN IF NOT EXISTS assumptions JSONB DEFAULT '[]'::jsonb;
ALTER TABLE decision_revisions ADD COLUMN IF NOT EXISTS evidence JSONB DEFAULT '[]'::jsonb;
ALTER TABLE decision_revisions ADD COLUMN IF NOT EXISTS alternatives JSONB DEFAULT '[]'::jsonb;
ALTER TABLE decision_revisions ADD COLUMN IF NOT EXISTS tags JSONB DEFAULT '[]'::jsonb;
ALTER TABLE decision_revisions ADD COLUMN IF NOT EXISTS canonical_json JSONB DEFAULT '{}'::jsonb;
ALTER TABLE decision_revisions ADD COLUMN IF NOT EXISTS payload_hash TEXT;
ALTER TABLE decision_revisions ADD COLUMN IF NOT EXISTS decision_hash BYTEA;
ALTER TABLE decision_revisions ADD COLUMN IF NOT EXISTS previous_revision_hash BYTEA;
ALTER TABLE decision_revisions ADD COLUMN IF NOT EXISTS author TEXT DEFAULT 'garuda-cli';
ALTER TABLE decision_revisions ADD COLUMN IF NOT EXISTS actor TEXT DEFAULT 'garuda-cli';
ALTER TABLE decision_revisions ADD COLUMN IF NOT EXISTS valid_from TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE decision_revisions ADD COLUMN IF NOT EXISTS valid_to TIMESTAMPTZ;
ALTER TABLE decision_revisions ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();