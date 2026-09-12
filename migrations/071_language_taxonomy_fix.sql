-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- Corrects mislabeled class entities.
--
-- Before this migration, the Python and TypeScript analyzers both
-- mapped AST "class" declarations to entity kind "struct", because no
-- KindClass constant existed. Neither language has structs. The data
-- was truthful about the source (a class is a class) but wrong about
-- the label.
--
-- This migration relabels existing rows. Future analyses use the new
-- KindClass constant. No structural change.

BEGIN;

UPDATE entities
SET kind = 'class'
WHERE kind = 'struct'
  AND language IN ('python', 'typescript');

-- Record the count so this migration's impact is auditable
DO $$
DECLARE
  affected integer;
BEGIN
  SELECT COUNT(*) INTO affected
  FROM entities
  WHERE kind = 'class'
    AND language IN ('python', 'typescript');

  RAISE NOTICE 'Relabeled % entities to kind=class', affected;
END $$;

COMMIT;