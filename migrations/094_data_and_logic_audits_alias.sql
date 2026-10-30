-- ==========================================================================
-- 094_data_and_logic_audits_alias.sql
-- Garuda v1.9.3: Data and Logic Audits View & Terminology Alignment
-- Re-categorizes financial audit tables into enterprise data & logic audits
-- to avoid fintech regulatory classification.
-- ==========================================================================

CREATE OR REPLACE VIEW data_audit_reconciliations AS
SELECT * FROM financial_audit_reconciliations;

