-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- The prior default (RESOLVED / AST_EXACT) claimed certainty for rows
-- that have no resolution classification. That contradicts Law 7.
-- A missing classification must default to the honest floor: AMBIGUOUS.
--
-- Every row currently holding the migration default is reset to the
-- honest floor. Real classifications will overwrite them on the next
-- analysis run of each repository.

ALTER TABLE claims
  ALTER COLUMN resolution_status SET DEFAULT 'AMBIGUOUS',
  ALTER COLUMN resolution_method SET DEFAULT 'HEURISTIC';

-- Reset rows that still hold the old default. Rows with real values
-- (GO_TYPES, IMPORT_RESOLUTION, HEURISTIC from a real write) are left alone
-- only if their method implies a real classification. Since we can't yet
-- distinguish, this resets everything written before the CLI was rebuilt.
UPDATE claims
   SET resolution_status = 'AMBIGUOUS',
       resolution_method = 'HEURISTIC'
 WHERE resolution_status = 'RESOLVED'
   AND resolution_method = 'AST_EXACT'
   AND confidence = 1.0;