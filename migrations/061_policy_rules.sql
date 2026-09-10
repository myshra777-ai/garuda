-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0

-- Extend existing policies table with rule storage
ALTER TABLE policies ADD COLUMN IF NOT EXISTS policy_version VARCHAR(32) DEFAULT 'v1';
ALTER TABLE policies ADD COLUMN IF NOT EXISTS rules JSONB DEFAULT '[]'::jsonb;
ALTER TABLE policies ADD COLUMN IF NOT EXISTS language VARCHAR(32) DEFAULT 'any';
ALTER TABLE policies ADD COLUMN IF NOT EXISTS priority INTEGER DEFAULT 100;
ALTER TABLE policies ADD COLUMN IF NOT EXISTS source_path TEXT DEFAULT '';
ALTER TABLE policies ADD COLUMN IF NOT EXISTS source_hash VARCHAR(64) DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_policies_language ON policies (tenant_id, language) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_policies_source_path ON policies (tenant_id, source_path);