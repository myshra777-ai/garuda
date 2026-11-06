-- 101: add language resolution methods to claims_resolution_method_check constraint.
DO $$
BEGIN
    IF to_regclass('public.claims') IS NOT NULL THEN
        ALTER TABLE claims DROP CONSTRAINT IF EXISTS claims_resolution_method_check;
        ALTER TABLE claims ADD CONSTRAINT claims_resolution_method_check
            CHECK (resolution_method IN (
                'GO_TYPES',
                'IMPORT_RESOLUTION',
                'AST_EXACT',
                'HEURISTIC',
                'GRAPH_DERIVED',
                'PYTHON_IMPORT',
                'TS_IMPORT',
                'JAVA_IMPORT',
                'JVM_IMPORT'
            ));
    END IF;
END $$;
