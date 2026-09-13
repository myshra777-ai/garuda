-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- 080_users_last_login_disabled.sql
--
-- Two columns on users.
--
-- last_login_at: written by the login handler on every successful
-- sign-in. Makes active-user counts computable from an indexed
-- column without scanning session tables.
--
-- is_disabled: revokes access without deleting the row. Audit
-- records, membership rows, and decisions reference users.id; a
-- deleted user leaves dangling references that would otherwise need
-- ON DELETE SET NULL across half the schema. Disabling is the
-- reversible alternative.

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMPTZ;

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS is_disabled BOOLEAN NOT NULL DEFAULT FALSE;

-- Index for the active-users aggregate query. Both the metric and
-- the middleware (Session C) read this column frequently.
CREATE INDEX IF NOT EXISTS idx_users_last_login
    ON users (last_login_at)
    WHERE last_login_at IS NOT NULL;

-- Index for the middleware's "is this user disabled?" check. Partial
-- index: only disabled users are indexed, because the middleware's
-- common case is "not disabled" and a full index would be a wasted
-- write per login.
CREATE INDEX IF NOT EXISTS idx_users_disabled
    ON users (id)
    WHERE is_disabled = TRUE;