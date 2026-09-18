-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- 088_mcp_sessions.sql
--
-- Session and tool-call tracking for the MCP server, plus the
-- observability tables the Control Plane reads. See:
--   docs/specs/mcp-sessions.md
--   docs/specs/control-plane.md
--
-- Creates five tables and one seed row:
--
--   mcp_sessions            — one row per MCP client process run
--   mcp_tool_calls          — one row per tools/call invocation
--   job_state               — single-row-per-job tracking for the
--                             pruning and stale-session jobs
--   control_tenant_notes    — private notes on a tenant, one row per
--                             tenant, visible only from the Control Plane
--   control_plane_access    — audit trail for the Control Plane
--
-- Plus the orphaned-session sentinel row in mcp_sessions, so that a
-- tools/call arriving after a failed session insert still records.
--
-- Every CREATE is IF NOT EXISTS. Re-applying this migration is a
-- no-op. The orphaned-session seed uses ON CONFLICT DO NOTHING.
--
-- Run inside one transaction. If any statement raises, nothing applies.

BEGIN;

-- ─────────────────────────────────────────────────────────────────────────────
-- 1. mcp_sessions
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS mcp_sessions (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id         UUID NOT NULL,
  workspace_id      UUID,
  session_token     TEXT NOT NULL,
  client_name       TEXT NOT NULL,
  client_version    TEXT,
  agent_id          TEXT,
  started_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  last_activity_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  closed_at         TIMESTAMPTZ,
  close_reason      TEXT,
  request_count     INT NOT NULL DEFAULT 0,
  tool_call_count   INT NOT NULL DEFAULT 0,
  CONSTRAINT mcp_sessions_close_reason_check CHECK (
    close_reason IS NULL
    OR close_reason IN ('graceful', 'stale', 'crashed', 'reactivated', 'orphaned')
  )
);

-- Active sessions for a tenant, newest first. Partial index on the
-- open-session set — closed sessions never appear in this query.
CREATE INDEX IF NOT EXISTS mcp_sessions_tenant_active_idx
  ON mcp_sessions (tenant_id, last_activity_at DESC)
  WHERE closed_at IS NULL;

-- Retention-curve queries scan by start time across all clients.
CREATE INDEX IF NOT EXISTS mcp_sessions_started_idx
  ON mcp_sessions (started_at DESC);

-- The stale-session job scans for open sessions with old activity.
-- Without this partial index, the job is a full table scan every
-- five minutes.
CREATE INDEX IF NOT EXISTS mcp_sessions_stale_idx
  ON mcp_sessions (last_activity_at)
  WHERE closed_at IS NULL;

-- ─────────────────────────────────────────────────────────────────────────────
-- 2. mcp_tool_calls
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS mcp_tool_calls (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id     UUID NOT NULL REFERENCES mcp_sessions(id) ON DELETE CASCADE,
  tool_name      TEXT NOT NULL,
  duration_ms    INT NOT NULL,
  status         TEXT NOT NULL,
  error_message  TEXT,
  called_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  args_summary   JSONB,
  CONSTRAINT mcp_tool_calls_status_check CHECK (status IN ('ok', 'error'))
);

-- Per-session tool-call history for the session drawer.
CREATE INDEX IF NOT EXISTS mcp_tool_calls_session_idx
  ON mcp_tool_calls (session_id, called_at DESC);

-- Per-tool error rate queries on the Operations tab.
CREATE INDEX IF NOT EXISTS mcp_tool_calls_tool_idx
  ON mcp_tool_calls (tool_name, called_at DESC);

-- The Operations tab's "errors in the last hour" panel runs on every
-- page load. The partial index on the error subset keeps it cheap
-- even as the table grows.
CREATE INDEX IF NOT EXISTS mcp_tool_calls_recent_idx
  ON mcp_tool_calls (called_at DESC)
  WHERE status = 'error';

-- ─────────────────────────────────────────────────────────────────────────────
-- 3. job_state
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS job_state (
  job_name         TEXT PRIMARY KEY,
  last_run_at      TIMESTAMPTZ NOT NULL,
  last_run_status  TEXT NOT NULL,
  last_error       TEXT,
  CONSTRAINT job_state_status_check CHECK (last_run_status IN ('ok', 'error'))
);

-- ─────────────────────────────────────────────────────────────────────────────
-- 4. control_tenant_notes
-- ─────────────────────────────────────────────────────────────────────────────
--
-- One row per tenant. A note is a field, not a log. If version history
-- is wanted later, it becomes a second table that references this one.
--
-- The tenant_id is the primary key, so a tenant has at most one note.
-- The FK to tenants ensures notes cannot reference a tenant that does
-- not exist and are removed if the tenant is deleted.

CREATE TABLE IF NOT EXISTS control_tenant_notes (
  tenant_id   UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
  body        TEXT NOT NULL DEFAULT '',
  updated_by  TEXT NOT NULL,
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─────────────────────────────────────────────────────────────────────────────
-- 5. control_plane_access
-- ─────────────────────────────────────────────────────────────────────────────
--
-- Audit trail for Control Plane access. Rows are written on auth events
-- (login attempts, failures, rate-limited requests) and on mutating
-- requests. Read-only GETs to /_/control/api/* do not write rows; the
-- volume would be high and the signal low.

CREATE TABLE IF NOT EXISTS control_plane_access (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  path         TEXT NOT NULL,
  method       TEXT NOT NULL,
  status       INT NOT NULL,
  source_ip    INET,
  user_agent   TEXT,
  auth_result  TEXT NOT NULL,
  occurred_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT control_plane_access_auth_result_check CHECK (
    auth_result IN ('ok', 'fail', 'rate_limited')
  )
);

-- The retention job prunes by occurred_at. The Operations tab reads
-- the last 24 hours.
CREATE INDEX IF NOT EXISTS control_plane_access_occurred_at_idx
  ON control_plane_access (occurred_at DESC);

-- ─────────────────────────────────────────────────────────────────────────────
-- 6. Orphaned-session sentinel
-- ─────────────────────────────────────────────────────────────────────────────
--
-- When the MCP server cannot create a session row (database
-- unreachable at startup, transient write failure), subsequent
-- tool-call rows have no valid session_id. Dropping those calls would
-- lose observability for exactly the sessions most worth observing.
--
-- This sentinel session absorbs them. Every Operations query filters
-- it out with WHERE client_name != 'orphaned'. The row itself is
-- visible if you look for it and invisible otherwise.
--
-- The tenant_id is the canonical tenant. mcp_sessions has no FK to
-- tenants, so a deployment that has not created the canonical tenant
-- yet can still seed this row.

INSERT INTO mcp_sessions (
  id, tenant_id, session_token, client_name, agent_id,
  started_at, last_activity_at, close_reason
) VALUES (
  '00000000-0000-0000-0000-000000000000',
  '00000000-0000-0000-0000-000000000001',
  'orphaned-session',
  'orphaned',
  'orphaned',
  NOW(), NOW(), 'orphaned'
) ON CONFLICT (id) DO NOTHING;

COMMIT;