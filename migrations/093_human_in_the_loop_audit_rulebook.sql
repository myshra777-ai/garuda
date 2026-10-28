-- ==========================================================================
-- 093_human_in_the_loop_audit_rulebook.sql
-- Garuda v1.9.0: Human-in-the-Loop Weighted Logic Disambiguation,
-- Surface Criticality Tiers, Financial/Data Audit Reconciliation,
-- and Auto-Generated Base Rulebook Approval
-- ==========================================================================

-- 1. Surface Criticality Tiers (Amazon-style Core vs Non-Blocking vs Minor UI)
CREATE TABLE IF NOT EXISTS surface_criticality_tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,
    workspace_id UUID,
    surface_name TEXT NOT NULL,
    package_pattern TEXT NOT NULL,
    symbol_pattern TEXT DEFAULT '.*',
    criticality_tier TEXT NOT NULL DEFAULT 'HIGH_SEVERE', -- HIGH_SEVERE | MEDIUM | LOW
    blast_radius_impact TEXT NOT NULL DEFAULT '',
    human_policy_action TEXT NOT NULL DEFAULT 'HUMAN_FREEZE_DO_NOT_TOUCH', -- HUMAN_FREEZE_DO_NOT_TOUCH | HUMAN_TRADE_OFF_REVIEW | ALLOW_WITH_INFO
    defined_by TEXT DEFAULT 'architect',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_surface_criticality_ws_name
    ON surface_criticality_tiers (workspace_id, surface_name);

-- 2. Multi-Team Weighted Logic Disambiguation Registry (0.0 to 1.0 Weights & Human Call)
CREATE TABLE IF NOT EXISTS logic_variants_registry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,
    workspace_id UUID,
    concept_name TEXT NOT NULL,           -- e.g. 'Loss Ratio Calculation'
    team_scope TEXT NOT NULL,             -- e.g. 'ops' | 'marketing' | 'management'
    variant_id TEXT NOT NULL,             -- e.g. 'loss_ratio_ops_v1'
    symbol_name TEXT NOT NULL,            -- e.g. 'ops.CalculateLossRatio'
    file_path TEXT NOT NULL,              -- e.g. 'internal/ops/loss_ratio.go:42'
    formula_expression TEXT NOT NULL,     -- e.g. '(Incurred_Claims + LAE) / Earned_Premium'
    calculated_weight NUMERIC(4,2) NOT NULL DEFAULT 0.70, -- 0.00 to 1.00
    severity_tier TEXT NOT NULL DEFAULT 'LOW',            -- HIGH_SEVERE | MEDIUM | LOW
    blast_radius_callers INT NOT NULL DEFAULT 0,
    blast_radius_surfaces TEXT NOT NULL DEFAULT '',
    human_impact_summary TEXT NOT NULL DEFAULT '',
    recommended_change_pct INT NOT NULL DEFAULT 70,
    human_decision_status TEXT NOT NULL DEFAULT 'AWAITING_HUMAN_DECISION', -- AWAITING_HUMAN_DECISION | APPROVED_BY_HUMAN | FROZEN_SEVERE | DEFERRED
    approved_by TEXT DEFAULT '',
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_logic_variants_ws_concept
    ON logic_variants_registry (workspace_id, concept_name);

-- 3. Financial & Business Data Audit Reconciliations (e.g., 48 Crore vs 4.8 Crore)
CREATE TABLE IF NOT EXISTS financial_audit_reconciliations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,
    workspace_id UUID,
    metric_name TEXT NOT NULL,            -- e.g. 'Total Claims Paid (FY26 Q3)'
    team_a_name TEXT NOT NULL,            -- e.g. 'Claims Ops & Settlement Team'
    team_a_value_raw NUMERIC(20,4) NOT NULL,
    team_a_value_formatted TEXT NOT NULL, -- e.g. '48.00 Crore INR'
    team_a_source_symbol TEXT NOT NULL,   -- e.g. 'ops.AggregateClaimsPaidCrore'
    team_b_name TEXT NOT NULL,            -- e.g. 'Finance & Management Reporting Team'
    team_b_value_raw NUMERIC(20,4) NOT NULL,
    team_b_value_formatted TEXT NOT NULL, -- e.g. '4.80 Crore INR'
    team_b_source_symbol TEXT NOT NULL,   -- e.g. 'finance.ExportExecutiveClaimsSummary'
    variance_ratio NUMERIC(12,4) NOT NULL DEFAULT 10.0,
    discrepancy_type TEXT NOT NULL DEFAULT 'SCALE_FACTOR_10X_MISMATCH',
    root_cause_ast_trace TEXT NOT NULL DEFAULT '',
    severity_tier TEXT NOT NULL DEFAULT 'HIGH_SEVERE',
    audit_status TEXT NOT NULL DEFAULT 'FLAGGED_FOR_HUMAN_AUDIT', -- FLAGGED_FOR_HUMAN_AUDIT | RECONCILED_BY_HUMAN
    reconciled_by TEXT DEFAULT '',
    reconciled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_financial_audit_ws
    ON financial_audit_reconciliations (workspace_id, audit_status);

-- 4. Auto-Synthesized Base Rulebooks (Human Review, Edit & 1-Click Approval)
CREATE TABLE IF NOT EXISTS synthesized_rulebooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,
    workspace_id UUID,
    version_tag TEXT NOT NULL DEFAULT 'v1.0-draft',
    status TEXT NOT NULL DEFAULT 'DRAFT_AWAITING_APPROVAL', -- DRAFT_AWAITING_APPROVAL | APPROVED_BASE_RULEBOOK
    title TEXT NOT NULL DEFAULT 'Auto-Synthesized Base Architecture & Business Rulebook',
    markdown_content TEXT NOT NULL DEFAULT '',
    rules_json JSONB DEFAULT '[]'::jsonb,
    approved_by TEXT DEFAULT '',
    approved_at TIMESTAMPTZ,
    merkle_hash TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_synthesized_rulebooks_ws
    ON synthesized_rulebooks (workspace_id, status);
