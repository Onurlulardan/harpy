-- +goose Up
-- +goose StatementBegin
DO $catalog_removal$
DECLARE
    object_record record;
    current_schema_name text := pg_catalog.current_schema();
    retired_at_value timestamp with time zone := pg_catalog.clock_timestamp();
BEGIN
    UPDATE __tenant_schema__.iam_role_permission_bindings AS binding
    SET state = 'revoked',
        revoked_at = retired_at_value,
        updated_at = retired_at_value,
        revision = binding.revision + 1
    FROM __tenant_schema__.iam_permissions AS permission
    JOIN __tenant_schema__.iam_resources AS resource
      ON resource.id = permission.resource_id
    WHERE binding.permission_id = permission.id
      AND resource.resource_key IN ('catalog', 'catalog.menu')
      AND binding.state = 'active';

    UPDATE __tenant_schema__.iam_user_permission_bindings AS binding
    SET state = 'revoked',
        revoked_at = retired_at_value,
        updated_at = retired_at_value,
        revision = binding.revision + 1
    FROM __tenant_schema__.iam_permissions AS permission
    JOIN __tenant_schema__.iam_resources AS resource
      ON resource.id = permission.resource_id
    WHERE binding.permission_id = permission.id
      AND resource.resource_key IN ('catalog', 'catalog.menu')
      AND binding.state = 'active';

    UPDATE __tenant_schema__.iam_organization_permission_bindings AS binding
    SET state = 'revoked',
        revoked_at = retired_at_value,
        updated_at = retired_at_value,
        revision = binding.revision + 1
    FROM __tenant_schema__.iam_permissions AS permission
    JOIN __tenant_schema__.iam_resources AS resource
      ON resource.id = permission.resource_id
    WHERE binding.permission_id = permission.id
      AND resource.resource_key IN ('catalog', 'catalog.menu')
      AND binding.state = 'active';

    UPDATE __tenant_schema__.iam_permissions AS permission
    SET state = 'retired',
        retired_at = retired_at_value,
        updated_at = retired_at_value,
        revision = permission.revision + 1
    FROM __tenant_schema__.iam_resources AS resource
    WHERE permission.resource_id = resource.id
      AND resource.resource_key IN ('catalog', 'catalog.menu')
      AND permission.state = 'active';

    UPDATE __tenant_schema__.iam_resources
    SET state = 'retired',
        retired_at = retired_at_value,
        updated_at = retired_at_value,
        revision = revision + 1
    WHERE resource_key IN ('catalog', 'catalog.menu')
      AND state = 'active';

    FOR object_record IN
        SELECT namespace.nspname, procedure_row.proname,
               pg_catalog.pg_get_function_identity_arguments(procedure_row.oid) AS identity_arguments
        FROM pg_catalog.pg_proc AS procedure_row
        JOIN pg_catalog.pg_namespace AS namespace
          ON namespace.oid = procedure_row.pronamespace
        WHERE namespace.nspname = current_schema_name
          AND procedure_row.proname ~ '^(catalog_|analytics_catalog_|iam_catalog_)'
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
          AND relation.relname ~ '^(catalog_|analytics_catalog_|iam_catalog_)'
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
          AND relation.relname ~ '^(catalog_|analytics_catalog_|iam_catalog_)'
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
          AND relation.relname ~ '^(catalog_|analytics_catalog_|iam_catalog_)'
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
          AND relation.relname ~ '^(catalog_|analytics_catalog_|iam_catalog_)'
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
          AND type_row.typname ~ '^(catalog_|analytics_catalog_|iam_catalog_)'
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
          AND type_row.typname ~ '^(catalog_|analytics_catalog_|iam_catalog_)'
    LOOP
        EXECUTE pg_catalog.format(
            'DROP TYPE IF EXISTS %I.%I CASCADE',
            object_record.nspname, object_record.typname
        );
    END LOOP;
END
$catalog_removal$;
-- +goose StatementEnd
