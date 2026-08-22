-- Migration 046: Ensure all decision_revisions columns exist

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