-- +goose Up
-- +goose StatementBegin
DO $catalog_removal$
DECLARE
    object_record record;
    current_schema_name text := pg_catalog.current_schema();
    retired_at_value timestamp with time zone := pg_catalog.clock_timestamp();
    has_rows boolean;
    unexpected_rows bigint;
BEGIN
    FOR object_record IN
        SELECT namespace.nspname, relation.relname
        FROM pg_catalog.pg_class AS relation
        JOIN pg_catalog.pg_namespace AS namespace
          ON namespace.oid = relation.relnamespace
        WHERE namespace.nspname = current_schema_name
          AND relation.relkind IN ('r', 'p')
          AND relation.relname ~ '^(catalog_|analytics_catalog_|iam_catalog_)'
    LOOP
        IF object_record.relname = 'catalog_menu_definitions' THEN
            EXECUTE pg_catalog.format(
                'SELECT count(*) FROM %I.%I WHERE definition_id <> $1 OR revision <> 1 OR active_version_id <> $2',
                object_record.nspname, object_record.relname
            ) INTO unexpected_rows
            USING '018f6000-0000-7000-8000-000000000001'::uuid,
                  '018f6000-0000-7000-8000-000000000002'::uuid;
            IF unexpected_rows > 0 THEN
                RAISE EXCEPTION USING
                    ERRCODE = '55000',
                    MESSAGE = 'Catalog removal requires an explicit data disposition for catalog_menu_definitions';
            END IF;
        ELSIF object_record.relname = 'catalog_menu_versions' THEN
            EXECUTE pg_catalog.format(
                'SELECT count(*) FROM %I.%I WHERE version_id <> $1 OR definition_id <> $2 OR version_number <> 1 OR content_sha256 <> $3',
                object_record.nspname, object_record.relname
            ) INTO unexpected_rows
            USING '018f6000-0000-7000-8000-000000000002'::uuid,
                  '018f6000-0000-7000-8000-000000000001'::uuid,
                  'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855';
            IF unexpected_rows > 0 THEN
                RAISE EXCEPTION USING
                    ERRCODE = '55000',
                    MESSAGE = 'Catalog removal requires an explicit data disposition for catalog_menu_versions';
            END IF;
        ELSE
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
          AND referenced_relation.relname ~ '^(catalog_|analytics_catalog_|iam_catalog_)'
          AND NOT (
              source_namespace.nspname = current_schema_name
              AND source_relation.relname ~ '^(catalog_|analytics_catalog_|iam_catalog_)'
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
          AND referenced_relation.relname ~ '^(catalog_|analytics_catalog_|iam_catalog_)'
          AND NOT (
              dependent_namespace.nspname = current_schema_name
              AND dependent_view.relname ~ '^(catalog_|analytics_catalog_|iam_catalog_)'
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
          AND referenced_relation.relname ~ '^(catalog_|analytics_catalog_|iam_catalog_)'
          AND NOT (
              dependent_namespace.nspname = current_schema_name
              AND dependent_routine.proname ~ '^(catalog_|analytics_catalog_|iam_catalog_)'
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
          AND referenced_routine.proname ~ '^(catalog_|analytics_catalog_|iam_catalog_)'
          AND NOT (
              dependent_namespace.nspname = current_schema_name
              AND dependent_routine.proname ~ '^(catalog_|analytics_catalog_|iam_catalog_)'
          )
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000',
            MESSAGE = 'Catalog removal blocked by a dependent routine outside the Catalog module';
    END IF;

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
          AND referenced_routine.proname ~ '^(catalog_|analytics_catalog_|iam_catalog_)'
          AND NOT (
              dependent_namespace.nspname = current_schema_name
              AND dependent_view.relname ~ '^(catalog_|analytics_catalog_|iam_catalog_)'
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
