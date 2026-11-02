-- +goose Up
-- +goose StatementBegin
DO $catalog_removal$
DECLARE
    object_record record;
    current_schema_name text := current_schema;
    has_rows boolean;
BEGIN
    FOR object_record IN
        SELECT namespace.nspname, relation.relname
        FROM pg_catalog.pg_class AS relation
        JOIN pg_catalog.pg_namespace AS namespace
          ON namespace.oid = relation.relnamespace
        WHERE namespace.nspname = current_schema_name
          AND relation.relkind IN ('r', 'p')
          AND relation.relname ~ '^catalog_'
    LOOP
        EXECUTE pg_catalog.format(
            'SELECT EXISTS (SELECT 1 FROM %I.%I)',
            object_record.nspname, object_record.relname
        ) INTO has_rows;
        IF has_rows THEN
            RAISE EXCEPTION USING
                ERRCODE = '55000',
                MESSAGE = pg_catalog.format(
                    'Catalog removal requires an explicit data disposition for %I.%I',
                    object_record.nspname, object_record.relname
                );
        END IF;
    END LOOP;

    IF EXISTS (
        SELECT 1
        FROM pg_catalog.pg_constraint AS foreign_key
        JOIN pg_catalog.pg_class AS referenced_relation
          ON referenced_relation.oid = foreign_key.confrelid
        JOIN pg_catalog.pg_namespace AS referenced_namespace
          ON referenced_namespace.oid = referenced_relation.relnamespace
        JOIN pg_catalog.pg_class AS source_relation
          ON source_relation.oid = foreign_key.conrelid
        JOIN pg_catalog.pg_namespace AS source_namespace
          ON source_namespace.oid = source_relation.relnamespace
        WHERE foreign_key.contype = 'f'
          AND referenced_namespace.nspname = current_schema_name
          AND referenced_relation.relname ~ '^catalog_'
          AND NOT (
              source_namespace.nspname = current_schema_name
              AND source_relation.relname ~ '^catalog_'
          )
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000',
            MESSAGE = 'Catalog removal blocked by a foreign key owned outside the Catalog module';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM pg_catalog.pg_depend AS dependency
        JOIN pg_catalog.pg_class AS referenced_relation
          ON dependency.refclassid = 'pg_catalog.pg_class'::regclass
         AND dependency.refobjid = referenced_relation.oid
        JOIN pg_catalog.pg_namespace AS referenced_namespace
          ON referenced_namespace.oid = referenced_relation.relnamespace
        JOIN pg_catalog.pg_rewrite AS view_rule
          ON dependency.classid = 'pg_catalog.pg_rewrite'::regclass
         AND dependency.objid = view_rule.oid
        JOIN pg_catalog.pg_class AS dependent_view
          ON dependent_view.oid = view_rule.ev_class
        JOIN pg_catalog.pg_namespace AS dependent_namespace
          ON dependent_namespace.oid = dependent_view.relnamespace
        WHERE dependency.deptype = 'n'
          AND referenced_namespace.nspname = current_schema_name
          AND referenced_relation.relname ~ '^catalog_'
          AND NOT (
              dependent_namespace.nspname = current_schema_name
              AND dependent_view.relname ~ '^catalog_'
          )
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000',
            MESSAGE = 'Catalog removal blocked by a view owned outside the Catalog module';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM pg_catalog.pg_depend AS dependency
        JOIN pg_catalog.pg_class AS referenced_relation
          ON dependency.refclassid = 'pg_catalog.pg_class'::regclass
         AND dependency.refobjid = referenced_relation.oid
        JOIN pg_catalog.pg_namespace AS referenced_namespace
          ON referenced_namespace.oid = referenced_relation.relnamespace
        JOIN pg_catalog.pg_proc AS dependent_routine
          ON dependency.classid = 'pg_catalog.pg_proc'::regclass
         AND dependency.objid = dependent_routine.oid
        JOIN pg_catalog.pg_namespace AS dependent_namespace
          ON dependent_namespace.oid = dependent_routine.pronamespace
        WHERE dependency.deptype = 'n'
          AND referenced_namespace.nspname = current_schema_name
          AND referenced_relation.relname ~ '^catalog_'
          AND NOT (
              dependent_namespace.nspname = current_schema_name
              AND dependent_routine.proname ~ '^catalog_'
          )
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000',
            MESSAGE = 'Catalog removal blocked by a routine owned outside the Catalog module';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM pg_catalog.pg_depend AS dependency
        JOIN pg_catalog.pg_proc AS referenced_routine
          ON dependency.refclassid = 'pg_catalog.pg_proc'::regclass
         AND dependency.refobjid = referenced_routine.oid
        JOIN pg_catalog.pg_namespace AS referenced_namespace
          ON referenced_namespace.oid = referenced_routine.pronamespace
        JOIN pg_catalog.pg_proc AS dependent_routine
          ON dependency.classid = 'pg_catalog.pg_proc'::regclass
         AND dependency.objid = dependent_routine.oid
        JOIN pg_catalog.pg_namespace AS dependent_namespace
          ON dependent_namespace.oid = dependent_routine.pronamespace
        WHERE dependency.deptype = 'n'
          AND referenced_namespace.nspname = current_schema_name
          AND referenced_routine.proname ~ '^catalog_'
          AND NOT (
              dependent_namespace.nspname = current_schema_name
              AND dependent_routine.proname ~ '^catalog_'
          )
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000',
            MESSAGE = 'Catalog removal blocked by a dependent routine outside the Catalog module';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM pg_catalog.pg_depend AS dependency
        JOIN pg_catalog.pg_proc AS referenced_routine
          ON dependency.refclassid = 'pg_catalog.pg_proc'::regclass
         AND dependency.refobjid = referenced_routine.oid
        JOIN pg_catalog.pg_namespace AS referenced_namespace
          ON referenced_namespace.oid = referenced_routine.pronamespace
        JOIN pg_catalog.pg_rewrite AS view_rule
          ON dependency.classid = 'pg_catalog.pg_rewrite'::regclass
         AND dependency.objid = view_rule.oid
        JOIN pg_catalog.pg_class AS dependent_view
          ON dependent_view.oid = view_rule.ev_class
        JOIN pg_catalog.pg_namespace AS dependent_namespace
          ON dependent_namespace.oid = dependent_view.relnamespace
        WHERE dependency.deptype = 'n'
          AND referenced_namespace.nspname = current_schema_name
          AND referenced_routine.proname ~ '^catalog_'
          AND NOT (
              dependent_namespace.nspname = current_schema_name
              AND dependent_view.relname ~ '^catalog_'
          )
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000',
            MESSAGE = 'Catalog removal blocked by a view dependent on a Catalog routine';
    END IF;

    FOR object_record IN
        SELECT namespace.nspname, procedure_row.proname,
               pg_catalog.pg_get_function_identity_arguments(procedure_row.oid) AS identity_arguments
        FROM pg_catalog.pg_proc AS procedure_row
        JOIN pg_catalog.pg_namespace AS namespace
          ON namespace.oid = procedure_row.pronamespace
        WHERE namespace.nspname = current_schema_name
          AND procedure_row.proname ~ '^catalog_'
          AND procedure_row.prokind IN ('f', 'p', 'a', 'w')
        ORDER BY procedure_row.oid DESC
    LOOP
        EXECUTE pg_catalog.format(
            'DROP ROUTINE IF EXISTS %I.%I(%s) CASCADE',
            object_record.nspname,
            object_record.proname,
            object_record.identity_arguments
        );
    END LOOP;

    FOR object_record IN
        SELECT namespace.nspname, relation.relname
        FROM pg_catalog.pg_class AS relation
        JOIN pg_catalog.pg_namespace AS namespace
          ON namespace.oid = relation.relnamespace
        WHERE namespace.nspname = current_schema_name
          AND relation.relkind = 'v'
          AND relation.relname ~ '^catalog_'
    LOOP
        EXECUTE pg_catalog.format(
            'DROP VIEW IF EXISTS %I.%I CASCADE',
            object_record.nspname, object_record.relname
        );
    END LOOP;

    FOR object_record IN
        SELECT namespace.nspname, relation.relname
        FROM pg_catalog.pg_class AS relation
        JOIN pg_catalog.pg_namespace AS namespace
          ON namespace.oid = relation.relnamespace
        WHERE namespace.nspname = current_schema_name
          AND relation.relkind = 'm'
          AND relation.relname ~ '^catalog_'
    LOOP
        EXECUTE pg_catalog.format(
            'DROP MATERIALIZED VIEW IF EXISTS %I.%I CASCADE',
            object_record.nspname, object_record.relname
        );
    END LOOP;

    FOR object_record IN
        SELECT namespace.nspname, relation.relname
        FROM pg_catalog.pg_class AS relation
        JOIN pg_catalog.pg_namespace AS namespace
          ON namespace.oid = relation.relnamespace
        WHERE namespace.nspname = current_schema_name
          AND relation.relkind IN ('r', 'p')
          AND relation.relname ~ '^catalog_'
    LOOP
        EXECUTE pg_catalog.format(
            'DROP TABLE IF EXISTS %I.%I CASCADE',
            object_record.nspname, object_record.relname
        );
    END LOOP;

    FOR object_record IN
        SELECT namespace.nspname, relation.relname
        FROM pg_catalog.pg_class AS relation
        JOIN pg_catalog.pg_namespace AS namespace
          ON namespace.oid = relation.relnamespace
        WHERE namespace.nspname = current_schema_name
          AND relation.relkind = 'S'
          AND relation.relname ~ '^catalog_'
    LOOP
        EXECUTE pg_catalog.format(
            'DROP SEQUENCE IF EXISTS %I.%I CASCADE',
            object_record.nspname, object_record.relname
        );
    END LOOP;

    FOR object_record IN
        SELECT namespace.nspname, type_row.typname
        FROM pg_catalog.pg_type AS type_row
        JOIN pg_catalog.pg_namespace AS namespace
          ON namespace.oid = type_row.typnamespace
        WHERE namespace.nspname = current_schema_name
          AND type_row.typrelid = 0
          AND type_row.typtype = 'd'
          AND type_row.typname ~ '^catalog_'
    LOOP
        EXECUTE pg_catalog.format(
            'DROP DOMAIN IF EXISTS %I.%I CASCADE',
            object_record.nspname, object_record.typname
        );
    END LOOP;

    FOR object_record IN
        SELECT namespace.nspname, type_row.typname
        FROM pg_catalog.pg_type AS type_row
        JOIN pg_catalog.pg_namespace AS namespace
          ON namespace.oid = type_row.typnamespace
        WHERE namespace.nspname = current_schema_name
          AND type_row.typrelid = 0
          AND type_row.typtype IN ('c', 'e', 'r', 'm')
          AND type_row.typname ~ '^catalog_'
    LOOP
        EXECUTE pg_catalog.format(
            'DROP TYPE IF EXISTS %I.%I CASCADE',
            object_record.nspname, object_record.typname
        );
    END LOOP;
END
$catalog_removal$;
-- +goose StatementEnd
