-- 040_add_canonical_identity_fields.sql
-- Add canonical identity fields to entities table

-- Add missing columns
ALTER TABLE entities 
ADD COLUMN IF NOT EXISTS line INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS module_path TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS package_path TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS receiver_type TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS commit_sha TEXT DEFAULT '';

-- Create indexes for faster queries
CREATE INDEX IF NOT EXISTS idx_entities_module_path ON entities(module_path);
CREATE INDEX IF NOT EXISTS idx_entities_package_path ON entities(package_path);
CREATE INDEX IF NOT EXISTS idx_entities_receiver_type ON entities(receiver_type);

-- Update existing rows: set line = line_start where line = 0
UPDATE entities SET line = line_start WHERE line = 0 AND line_start > 0;

-- Add column for epistemic_class if not exists
ALTER TABLE claims 
ADD COLUMN IF NOT EXISTS epistemic_class TEXT DEFAULT 'OBSERVED_CODE',
ADD COLUMN IF NOT EXISTS confidence FLOAT DEFAULT 1.0;

-- Add index for epistemic_class
CREATE INDEX IF NOT EXISTS idx_claims_epistemic_class ON claims(epistemic_class);