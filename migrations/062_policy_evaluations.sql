-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS policy_evaluations (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id          UUID NOT NULL,
    workspace_id       UUID NOT NULL,
    policy_id          UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    policy_version     VARCHAR(32) NOT NULL,
    decision           VARCHAR(16) NOT NULL,    -- ALLOW | WARN | REVIEW | BLOCK
    reason             TEXT NOT NULL,
    evidence           JSONB NOT NULL DEFAULT '{}'::jsonb,
    evaluated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    merkle_block_height BIGINT,
    merkle_proof       BYTEA,
    subject_kind       VARCHAR(32) DEFAULT '',   -- commit | pr | agent_action | snapshot
    subject_id         TEXT DEFAULT '',
    actor              VARCHAR(128) DEFAULT '',
    CONSTRAINT policy_evaluations_decision_check
        CHECK (decision IN ('ALLOW', 'WARN', 'REVIEW', 'BLOCK'))
);

CREATE INDEX IF NOT EXISTS idx_policy_eval_workspace ON policy_evaluations (tenant_id, workspace_id, evaluated_at DESC);
CREATE INDEX IF NOT EXISTS idx_policy_eval_policy ON policy_evaluations (policy_id, evaluated_at DESC);
CREATE INDEX IF NOT EXISTS idx_policy_eval_decision ON policy_evaluations (workspace_id, decision);
CREATE INDEX IF NOT EXISTS idx_policy_eval_subject ON policy_evaluations (subject_kind, subject_id);

COMMENT ON TABLE policy_evaluations IS
'Immutable log of enforcement decisions. Each row is anchored to the Merkle ledger.';