-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- 086_mcp_agent_watermarks.sql
--
-- Tracks the last time each MCP agent was briefed, so garuda.briefing
-- can compute a diff since the caller's previous session.
--
-- Keyed on (tenant_id, workspace_id, agent_id). agent_id is derived
-- by the MCP server in this order:
--
--   1. args["agent_id"] — caller-supplied, the escape hatch
--   2. clientInfo.name from the MCP initialize handshake
--   3. "default"
--
-- session_id is not part of the key. MCP sessions are ephemeral;
-- keying on session_id would lose continuity on every client restart.
-- session_id is returned in the briefing response as a correlation
-- identifier, not used for the watermark.
--
-- last_briefed_at is the wall-clock time of the last briefing. It is
-- the primary filter for "what is new": claims.created_at and
-- claim_verifications.last_evaluated_at greater than this value.
--
-- last_merkle_height records the tenant's Merkle block height at the
-- moment of the last briefing. It allows a future version to report
-- "nothing has been anchored since your last session" without a
-- timestamp comparison.

BEGIN;

CREATE TABLE IF NOT EXISTS mcp_agent_watermarks (
    tenant_id           UUID NOT NULL,
    workspace_id        UUID NOT NULL,
    agent_id            TEXT NOT NULL,
    last_briefed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_merkle_height  BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (tenant_id, workspace_id, agent_id)
);

CREATE INDEX IF NOT EXISTS idx_agent_watermarks_briefed_at
    ON mcp_agent_watermarks (tenant_id, workspace_id, last_briefed_at DESC);

COMMIT;