-- +goose Up
-- +goose StatementBegin
DO $catalog_removal$
DECLARE
    object_record record;
    current_schema_name text := pg_catalog.current_schema();
BEGIN
    FOR object_record IN
        SELECT namespace.nspname, procedure_row.proname,
               pg_catalog.pg_get_function_identity_arguments(procedure_row.oid) AS identity_arguments
        FROM pg_catalog.pg_proc AS procedure_row
        JOIN pg_catalog.pg_namespace AS namespace
          ON namespace.oid = procedure_row.pronamespace
        WHERE namespace.nspname = current_schema_name
          AND procedure_row.proname ~ '^catalog_'
          AND procedure_row.prokind IN ('f', 'p', 'a')
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
