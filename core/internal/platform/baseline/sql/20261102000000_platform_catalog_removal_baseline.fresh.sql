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

    DECLARE
        removal_name_pattern text := '^catalog_';
        drop_kind text;
        drop_oid oid;
        drop_command text;
        remaining_objects integer;
        constraint_record record;
    BEGIN
        CREATE TEMP TABLE catalog_removal_targets (
            object_kind text NOT NULL,
            object_schema text NOT NULL,
            object_name text NOT NULL,
            identity_arguments text NOT NULL,
            object_oid oid NOT NULL,
            PRIMARY KEY (object_kind, object_oid)
        ) ON COMMIT DROP;

        INSERT INTO catalog_removal_targets (
            object_kind, object_schema, object_name, identity_arguments, object_oid
        )
        SELECT CASE relation.relkind
                WHEN 'v' THEN 'view'
                WHEN 'm' THEN 'materialized_view'
                WHEN 'S' THEN 'sequence'
                ELSE 'table'
            END,
            namespace.nspname, relation.relname, '', relation.oid
        FROM pg_catalog.pg_class AS relation
        JOIN pg_catalog.pg_namespace AS namespace
          ON namespace.oid = relation.relnamespace
        WHERE namespace.nspname = current_schema_name
          AND relation.relkind IN ('r', 'p', 'v', 'm', 'S')
          AND relation.relname ~ removal_name_pattern;

        INSERT INTO catalog_removal_targets (
            object_kind, object_schema, object_name, identity_arguments, object_oid
        )
        SELECT 'routine', namespace.nspname, procedure_row.proname,
            pg_catalog.pg_get_function_identity_arguments(procedure_row.oid),
            procedure_row.oid
        FROM pg_catalog.pg_proc AS procedure_row
        JOIN pg_catalog.pg_namespace AS namespace
          ON namespace.oid = procedure_row.pronamespace
        WHERE namespace.nspname = current_schema_name
          AND procedure_row.proname ~ removal_name_pattern
          AND procedure_row.prokind IN ('f', 'p', 'a', 'w');

        INSERT INTO catalog_removal_targets (
            object_kind, object_schema, object_name, identity_arguments, object_oid
        )
        SELECT CASE WHEN type_row.typtype = 'd' THEN 'domain' ELSE 'type' END,
            namespace.nspname, type_row.typname, '', type_row.oid
        FROM pg_catalog.pg_type AS type_row
        JOIN pg_catalog.pg_namespace AS namespace
          ON namespace.oid = type_row.typnamespace
        WHERE namespace.nspname = current_schema_name
          AND type_row.typrelid = 0
          AND type_row.typtype IN ('c', 'd', 'e', 'm', 'r')
          AND type_row.typname ~ removal_name_pattern;

        IF EXISTS (
            SELECT 1
            FROM pg_catalog.pg_trigger AS trigger_row
            JOIN pg_catalog.pg_class AS relation
              ON relation.oid = trigger_row.tgrelid
            JOIN pg_catalog.pg_namespace AS namespace
              ON namespace.oid = relation.relnamespace
            JOIN pg_catalog.pg_depend AS trigger_dependency
              ON trigger_dependency.classid = 'pg_catalog.pg_trigger'::regclass
             AND trigger_dependency.objid = trigger_row.oid
             AND trigger_dependency.deptype = 'n'
            JOIN catalog_removal_targets AS routine_target
              ON routine_target.object_kind = 'routine'
             AND routine_target.object_oid = trigger_dependency.refobjid
            WHERE NOT (
                namespace.nspname = current_schema_name
                AND relation.relname ~ removal_name_pattern
            )
              AND NOT trigger_row.tgisinternal
        ) THEN
            RAISE EXCEPTION USING
                ERRCODE = '55000',
                MESSAGE = 'Catalog removal blocked by a trigger owned outside the Catalog module';
        END IF;

        FOR constraint_record IN
            SELECT source_namespace.nspname AS source_schema,
                source_relation.relname AS source_table,
                foreign_key.conname AS constraint_name
            FROM pg_catalog.pg_constraint AS foreign_key
            JOIN catalog_removal_targets AS source_table
              ON source_table.object_kind = 'table'
             AND source_table.object_oid = foreign_key.conrelid
            JOIN catalog_removal_targets AS referenced_table
              ON referenced_table.object_kind = 'table'
             AND referenced_table.object_oid = foreign_key.confrelid
            JOIN pg_catalog.pg_class AS source_relation
              ON source_relation.oid = foreign_key.conrelid
            JOIN pg_catalog.pg_namespace AS source_namespace
              ON source_namespace.oid = source_relation.relnamespace
            WHERE foreign_key.contype = 'f'
        LOOP
            EXECUTE pg_catalog.format(
                'ALTER TABLE %I.%I DROP CONSTRAINT %I',
                constraint_record.source_schema,
                constraint_record.source_table,
                constraint_record.constraint_name
            );
        END LOOP;

        LOOP
            DELETE FROM catalog_removal_targets AS target
            WHERE target.object_kind IN ('view', 'materialized_view', 'table', 'sequence')
              AND NOT EXISTS (
                  SELECT 1 FROM pg_catalog.pg_class AS relation
                  WHERE relation.oid = target.object_oid
              );
            DELETE FROM catalog_removal_targets AS target
            WHERE target.object_kind = 'routine'
              AND NOT EXISTS (
                  SELECT 1 FROM pg_catalog.pg_proc AS procedure_row
                  WHERE procedure_row.oid = target.object_oid
              );
            DELETE FROM catalog_removal_targets AS target
            WHERE target.object_kind IN ('domain', 'type')
              AND NOT EXISTS (
                  SELECT 1 FROM pg_catalog.pg_type AS type_row
                  WHERE type_row.oid = target.object_oid
              );

            SELECT target.object_kind, target.object_oid,
                CASE target.object_kind
                    WHEN 'routine' THEN pg_catalog.format(
                        'DROP ROUTINE IF EXISTS %I.%I(%s)',
                        target.object_schema, target.object_name, target.identity_arguments)
                    WHEN 'view' THEN pg_catalog.format(
                        'DROP VIEW IF EXISTS %I.%I',
                        target.object_schema, target.object_name)
                    WHEN 'materialized_view' THEN pg_catalog.format(
                        'DROP MATERIALIZED VIEW IF EXISTS %I.%I',
                        target.object_schema, target.object_name)
                    WHEN 'table' THEN pg_catalog.format(
                        'DROP TABLE IF EXISTS %I.%I',
                        target.object_schema, target.object_name)
                    WHEN 'sequence' THEN pg_catalog.format(
                        'DROP SEQUENCE IF EXISTS %I.%I',
                        target.object_schema, target.object_name)
                    WHEN 'domain' THEN pg_catalog.format(
                        'DROP DOMAIN IF EXISTS %I.%I',
                        target.object_schema, target.object_name)
                    ELSE pg_catalog.format(
                        'DROP TYPE IF EXISTS %I.%I',
                        target.object_schema, target.object_name)
                END
            INTO drop_kind, drop_oid, drop_command
            FROM catalog_removal_targets AS target
            WHERE NOT EXISTS (
                SELECT 1
                FROM pg_catalog.pg_depend AS dependency
                WHERE dependency.refobjid = target.object_oid
                  AND dependency.deptype = 'n'
                  AND (
                      (
                          dependency.classid = 'pg_catalog.pg_proc'::regclass
                          AND EXISTS (
                              SELECT 1 FROM catalog_removal_targets AS other
                              WHERE other.object_kind = 'routine'
                                AND other.object_oid = dependency.objid
                                AND other.object_oid <> target.object_oid
                          )
                      )
                      OR (
                          dependency.classid = 'pg_catalog.pg_class'::regclass
                          AND EXISTS (
                              SELECT 1 FROM catalog_removal_targets AS other
                              WHERE other.object_kind IN (
                                  'view', 'materialized_view', 'table', 'sequence'
                              )
                                AND other.object_oid = dependency.objid
                                AND other.object_oid <> target.object_oid
                          )
                      )
                      OR (
                          dependency.classid = 'pg_catalog.pg_type'::regclass
                          AND EXISTS (
                              SELECT 1 FROM catalog_removal_targets AS other
                              WHERE other.object_kind IN ('domain', 'type')
                                AND other.object_oid = dependency.objid
                                AND other.object_oid <> target.object_oid
                          )
                      )
                      OR (
                          dependency.classid = 'pg_catalog.pg_rewrite'::regclass
                          AND EXISTS (
                              SELECT 1
                              FROM pg_catalog.pg_rewrite AS view_rule
                              JOIN catalog_removal_targets AS other
                                ON other.object_oid = view_rule.ev_class
                               AND other.object_kind IN ('view', 'materialized_view')
                              WHERE view_rule.oid = dependency.objid
                                AND other.object_oid <> target.object_oid
                          )
                      )
                  )
            )
            AND NOT (
                target.object_kind = 'sequence'
                AND EXISTS (
                    SELECT 1
                    FROM pg_catalog.pg_depend AS dependency
                    WHERE dependency.classid = 'pg_catalog.pg_class'::regclass
                      AND dependency.objid = target.object_oid
                      AND dependency.deptype IN ('a', 'i', 'n')
                      AND EXISTS (
                          SELECT 1 FROM catalog_removal_targets AS owner_table
                          WHERE owner_table.object_kind = 'table'
                            AND (
                                owner_table.object_oid = dependency.refobjid
                                OR EXISTS (
                                    SELECT 1
                                    FROM pg_catalog.pg_attribute AS attribute
                                    WHERE dependency.refclassid = 'pg_catalog.pg_attribute'::regclass
                                      AND dependency.refobjid = attribute.attrelid
                                      AND dependency.refobjsubid = attribute.attnum
                                      AND attribute.attrelid = owner_table.object_oid
                                )
                            )
                      )
                )
            )
            AND NOT EXISTS (
                SELECT 1
                FROM pg_catalog.pg_constraint AS foreign_key
                JOIN catalog_removal_targets AS other
                  ON other.object_oid = foreign_key.conrelid
                 AND other.object_kind = 'table'
                WHERE target.object_kind = 'table'
                  AND foreign_key.contype = 'f'
                  AND foreign_key.confrelid = target.object_oid
                  AND foreign_key.conrelid <> target.object_oid
            )
            AND NOT (
                target.object_kind = 'routine'
                AND EXISTS (
                    SELECT 1
                    FROM pg_catalog.pg_depend AS trigger_dependency
                    JOIN pg_catalog.pg_trigger AS trigger_row
                      ON trigger_dependency.classid = 'pg_catalog.pg_trigger'::regclass
                     AND trigger_dependency.objid = trigger_row.oid
                    JOIN catalog_removal_targets AS owner_table
                      ON owner_table.object_kind = 'table'
                     AND owner_table.object_oid = trigger_row.tgrelid
                    WHERE trigger_dependency.refobjid = target.object_oid
                      AND trigger_dependency.deptype = 'n'
                )
            )
            ORDER BY target.object_kind, target.object_name
            LIMIT 1;

            IF NOT FOUND THEN
                SELECT count(*) INTO remaining_objects FROM catalog_removal_targets;
                IF remaining_objects > 0 THEN
                    RAISE EXCEPTION USING
                        ERRCODE = '55000',
                        MESSAGE = 'Catalog removal blocked by an unresolved dependency inside the Catalog module';
                END IF;
                EXIT;
            END IF;

            EXECUTE drop_command;
            DELETE FROM catalog_removal_targets
            WHERE object_kind = drop_kind AND object_oid = drop_oid;
        END LOOP;
    END;
END
$catalog_removal$;
-- +goose StatementEnd
