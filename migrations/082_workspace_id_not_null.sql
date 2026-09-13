-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0

BEGIN;

-- Precondition: zero NULL workspace_id rows in any of the four tables.
-- Measured 2026-09-13: entities=0, claims=0, relationships=0, repositories=0.
DO $$
DECLARE
  n bigint;
BEGIN
  SELECT COUNT(*) INTO n FROM entities     WHERE workspace_id IS NULL;
  IF n > 0 THEN RAISE EXCEPTION 'entities has % NULL workspace_id rows', n; END IF;
  SELECT COUNT(*) INTO n FROM claims       WHERE workspace_id IS NULL;
  IF n > 0 THEN RAISE EXCEPTION 'claims has % NULL workspace_id rows', n; END IF;
  SELECT COUNT(*) INTO n FROM relationships WHERE workspace_id IS NULL;
  IF n > 0 THEN RAISE EXCEPTION 'relationships has % NULL workspace_id rows', n; END IF;
  SELECT COUNT(*) INTO n FROM repositories WHERE workspace_id IS NULL;
  IF n > 0 THEN RAISE EXCEPTION 'repositories has % NULL workspace_id rows', n; END IF;
END $$;

ALTER TABLE entities      ALTER COLUMN workspace_id SET NOT NULL;
ALTER TABLE claims        ALTER COLUMN workspace_id SET NOT NULL;
ALTER TABLE relationships ALTER COLUMN workspace_id SET NOT NULL;
ALTER TABLE repositories  ALTER COLUMN workspace_id SET NOT NULL;

COMMIT;