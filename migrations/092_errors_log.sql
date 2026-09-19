-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.
--
-- 092: errors_log
--
-- Durable record of WARN and ERROR level slog entries. Written by the
-- API server's async error log writer, read by the Control Plane
-- Operations tab. Every row carries a request_id when the emitting
-- site supplied one.

BEGIN;

CREATE TABLE IF NOT EXISTS errors_log (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    occurred_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    level        TEXT NOT NULL CHECK (level IN ('WARN', 'ERROR')),
    message      TEXT NOT NULL,
    request_id   TEXT,
    details      JSONB
);

CREATE INDEX IF NOT EXISTS errors_log_recent_idx
    ON errors_log (occurred_at DESC);

CREATE INDEX IF NOT EXISTS errors_log_level_recent_idx
    ON errors_log (level, occurred_at DESC);

-- Ensure the read-only Control Plane role can read this table.
-- Migration 089 set ALTER DEFAULT PRIVILEGES, which should cover it,
-- but a corrupt grant path should fail loudly rather than silently.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'garuda_control_ro') THEN
        EXECUTE 'GRANT SELECT ON errors_log TO garuda_control_ro';
    END IF;
END $$;

COMMIT;