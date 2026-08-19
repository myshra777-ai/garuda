-- Add line number columns to entities table
ALTER TABLE entities 
ADD COLUMN IF NOT EXISTS line_start INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS line_end INTEGER DEFAULT 0;

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_entities_line_start ON entities(line_start);
CREATE INDEX IF NOT EXISTS idx_entities_line_end ON entities(line_end);

-- Update existing rows with line numbers from line column
UPDATE entities SET line_start = line WHERE line_start = 0 AND line > 0;
UPDATE entities SET line_end = line WHERE line_end = 0 AND line > 0;

-- Add columns to evidence tracking if needed
ALTER TABLE evidence_store 
ADD COLUMN IF NOT EXISTS line_start INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS line_end INTEGER DEFAULT 0;