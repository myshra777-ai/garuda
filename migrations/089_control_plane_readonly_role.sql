-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- 089_control_plane_readonly_role.sql
--
-- Creates the read-only database role that the Control Plane uses.
-- See docs/specs/control-plane-metrics.md §Data access model.
--
-- The role has SELECT on every table in the public schema, plus
-- default privileges for tables created by future migrations. It
-- has no INSERT, UPDATE, DELETE, or TRUNCATE, even if a future
-- migration accidentally grants the default public role.
--
-- The password is NOT set by this migration. Anything the migration
-- writes is committed to git. The operator sets the password once,
-- from a terminal, before starting the API server with the
-- CONTROL_DATABASE_URL env var:
--
--   ALTER ROLE garuda_control_ro WITH PASSWORD '<generated>';
--
--   # generate with:
--   #   openssl rand -base64 32
--
--   export CONTROL_DATABASE_URL="postgres://garuda_control_ro:<generated>@localhost:5433/garuda_test?sslmode=disable"
--
-- If CONTROL_DATABASE_URL is unset, the Control Plane returns 404
-- for every request. Same treatment as an unset GARUDA_CONTROL_TOKEN.
--
-- Idempotent. Re-applying is a no-op.

BEGIN;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'garuda_control_ro') THEN
    CREATE ROLE garuda_control_ro
      WITH LOGIN
           NOSUPERUSER
           NOCREATEDB
           NOCREATEROLE
           NOINHERIT;
  END IF;
END $$;

-- SELECT on every existing table in the public schema.
GRANT SELECT ON ALL TABLES IN SCHEMA public TO garuda_control_ro;

-- SELECT on every future table. Applies to migrations run by the
-- current role after this point.
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT ON TABLES TO garuda_control_ro;

-- Explicit denial. The grants above only give SELECT, but this
-- makes the intent unmissable to a future reader.
REVOKE INSERT, UPDATE, DELETE, TRUNCATE
  ON ALL TABLES IN SCHEMA public
  FROM garuda_control_ro;

COMMIT;
