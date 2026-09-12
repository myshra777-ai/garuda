-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- Adds language-specific import resolution methods to the claims table
-- CHECK constraint.
--
-- Migration 063 defined the allowed resolution_method values as
-- GO_TYPES, IMPORT_RESOLUTION, AST_EXACT, HEURISTIC, GRAPH_DERIVED.
-- These were Go-centric. Path A (Python and TypeScript import
-- resolution) introduces two new methods:
--
--   PYTHON_IMPORT  — Python import statement resolved to a module
--                    or symbol within the workspace
--   TS_IMPORT      — TypeScript/JavaScript import resolved to a file
--                    within the workspace
--
-- Both are Tier 2 resolutions. Neither claims type information.
-- Both are honest lower-bound classifications compared to what a
-- full resolver (jedi, tsserver) would produce in Path B.

BEGIN;

ALTER TABLE claims
  DROP CONSTRAINT IF EXISTS claims_resolution_method_check;

ALTER TABLE claims
  ADD CONSTRAINT claims_resolution_method_check
  CHECK (resolution_method IN (
    'GO_TYPES',
    'IMPORT_RESOLUTION',
    'AST_EXACT',
    'HEURISTIC',
    'GRAPH_DERIVED',
    'PYTHON_IMPORT',
    'TS_IMPORT'
  ));

COMMIT;