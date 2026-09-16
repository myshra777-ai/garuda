-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- 087_policies_tenant_statement_uniq.sql
--
-- Adds a unique constraint on (tenant_id, statement). The policies
-- table had no unique constraint other than the primary key, so
-- upsertPolicy's select-then-write pattern had no DB-level guard.
-- Two concurrent evaluations could both find no existing row and both
-- insert, producing two active rows with the same statement.
--
-- Precondition: no duplicate (tenant_id, statement) rows. The
-- migration refuses if any exist.

BEGIN;

-- Guard: refuse if duplicates exist under the new key.
DO $$
DECLARE
  dup_count INT;
BEGIN
  SELECT COUNT(*) INTO dup_count FROM (
    SELECT 1 FROM policies
     GROUP BY tenant_id, statement
    HAVING COUNT(*) > 1
  ) d;
  IF dup_count > 0 THEN
    RAISE EXCEPTION '087: % duplicate (tenant_id, statement) groups exist; reconcile before applying', dup_count;
  END IF;
END $$;

ALTER TABLE policies
  ADD CONSTRAINT policies_tenant_statement_uniq
  UNIQUE (tenant_id, statement);

COMMIT;