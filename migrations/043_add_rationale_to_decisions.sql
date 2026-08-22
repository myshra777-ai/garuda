-- Migration 043: Ensure rationale column exists on decisions
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS rationale TEXT;