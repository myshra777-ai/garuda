-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- Adds explicit resolution classification to claims.
-- Every claim must declare HOW it was resolved (type-checked, import-matched, heuristic)
-- and WHETHER it is RESOLVED, AMBIGUOUS, UNRESOLVED, or INFERRED.
--
-- This is the schema companion to Law 7 (explicit uncertainty).
-- Existing rows default to RESOLVED/AST_EXACT — they were produced before this
-- column existed and their true classification is unknown. New writes must
-- populate both fields explicitly.

ALTER TABLE claims
  ADD COLUMN IF NOT EXISTS resolution_status text NOT NULL DEFAULT 'RESOLVED',
  ADD COLUMN IF NOT EXISTS resolution_method text NOT NULL DEFAULT 'AST_EXACT';

ALTER TABLE claims
  ADD CONSTRAINT claims_resolution_status_check
  CHECK (resolution_status IN ('RESOLVED','AMBIGUOUS','UNRESOLVED','INFERRED'));

ALTER TABLE claims
  ADD CONSTRAINT claims_resolution_method_check
  CHECK (resolution_method IN ('GO_TYPES','IMPORT_RESOLUTION','AST_EXACT','HEURISTIC','GRAPH_DERIVED'));

CREATE INDEX IF NOT EXISTS idx_claims_resolution
  ON claims(workspace_id, resolution_status);