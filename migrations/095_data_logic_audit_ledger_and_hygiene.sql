-- ==========================================================================
-- 095_data_logic_audit_ledger_and_hygiene.sql
-- Garuda v1.9.4:
-- 1. Renames/aliases Financial Ledger terminology to Data & Logic Audit Ledger
--    (data_logic_audit_ledger & data_audit_reconciliations) to avoid fintech
--    regulatory misclassification while preserving backward compatibility.
-- 2. Cleans all fabricated 'Human Stakeholder' approvals and smoke-test rows
--    (BUG-01 & BUG-21).
-- ==========================================================================

-- 1. Create canonical Data & Logic Audit Ledger views
CREATE OR REPLACE VIEW data_logic_audit_ledger AS
SELECT * FROM financial_audit_reconciliations;

CREATE OR REPLACE VIEW data_audit_reconciliations AS
SELECT * FROM financial_audit_reconciliations;

-- 2. Reset any pre-seeded or MCP-fabricated 'Human Stakeholder' variant approvals (BUG-01 / BUG-21)
UPDATE logic_variants_registry
   SET human_decision_status = CASE
           WHEN severity_tier = 'HIGH_SEVERE' THEN 'FROZEN_SEVERE_DO_NOT_TOUCH'
           ELSE 'AWAITING_HUMAN_DECISION'
       END,
       approved_by = '',
       approved_at = NULL,
       updated_at = NOW()
 WHERE approved_by ILIKE '%Human Stakeholder%'
    OR approved_by ILIKE '%unverified-mcp-call%'
    OR approved_by ILIKE '%submitted via mcp-client%'
    OR (human_decision_status = 'APPROVED_BY_HUMAN' AND COALESCE(approved_by, '') = '');

-- 3. Reset any unauthenticated MCP audit reconciliations (BUG-01 / BUG-21)
UPDATE financial_audit_reconciliations
   SET audit_status = 'FLAGGED_FOR_HUMAN_AUDIT',
       reconciled_by = '',
       reconciled_at = NULL
 WHERE reconciled_by ILIKE '%Human Stakeholder%'
    OR reconciled_by ILIKE '%mcp-client%'
    OR reconciled_by ILIKE '%mcp-agent%'
    OR (audit_status = 'RECONCILED_BY_HUMAN' AND COALESCE(reconciled_by, '') = '');

-- 4. Update surface criticality tier terminology from Financial/Claims Settlement Ledger to Data & Logic Audit Ledger
UPDATE surface_criticality_tiers
   SET surface_name = 'Core Data & Logic Audit Ledger & RFC 6962 Merkle Seal',
       blast_radius_impact = 'SEVERE BLAST RADIUS: Core data & logic audit ledger and cryptographic Merkle proof integrity (60 callers).',
       updated_at = NOW()
 WHERE surface_name ILIKE '%Claims Settlement Ledger%'
    OR surface_name ILIKE '%Settlement Ledger%';

-- 5. Delete smoke-test decision drafts (BUG-21)
DELETE FROM decisions
 WHERE (title ILIKE '%smoke%' OR statement ILIKE '%smoke%' OR owner = 'mcp-agent' OR owner LIKE 'mcp-client:%')
   AND status = 'draft';
