-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION cutover_authority.catalog_definition_sha256_v1(p_kind text, p_object_oid oid)
RETURNS text LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_located_count integer;
BEGIN
    IF p_kind IS NULL OR p_kind NOT IN ('schema','table','function') OR
       p_object_oid IS NULL OR p_object_oid = 0 THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'catalog definition identity is invalid';
    END IF;
    SELECT pg_catalog.count(*) INTO v_located_count FROM (
        SELECT namespace.oid FROM pg_catalog.pg_namespace AS namespace
        WHERE p_kind = 'schema' AND namespace.oid = p_object_oid
        UNION ALL
        SELECT relation.oid FROM pg_catalog.pg_class AS relation
        WHERE p_kind = 'table' AND relation.oid = p_object_oid
          AND relation.relkind IN ('r','p')
        UNION ALL
        SELECT procedure_row.oid FROM pg_catalog.pg_proc AS procedure_row
        WHERE p_kind = 'function' AND procedure_row.oid = p_object_oid
    ) AS located;
    IF v_located_count IS DISTINCT FROM 1 THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'catalog definition object is absent';
    END IF;
RETURN (SELECT pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
    pg_catalog.jsonb_build_object(
        'namespace', (SELECT pg_catalog.jsonb_build_object('name', namespace.nspname)
            FROM pg_catalog.pg_namespace AS namespace
            WHERE p_kind = 'schema' AND namespace.oid = p_object_oid),
        'relation', (SELECT pg_catalog.jsonb_build_object(
            'schema', namespace.nspname, 'name', relation.relname,
            'kind', relation.relkind, 'persistence', relation.relpersistence,
            'replica_identity', relation.relreplident,
            'access_method', access_method.amname, 'tablespace', tablespace.spcname,
            'options', COALESCE((SELECT pg_catalog.jsonb_agg(option_row ORDER BY option_row)
                FROM pg_catalog.unnest(relation.reloptions) AS option_row), '[]'::jsonb),
            'is_partition', relation.relispartition,
            'partition_key', pg_catalog.pg_get_partkeydef(relation.oid),
            'partition_bound', pg_catalog.pg_get_expr(relation.relpartbound, relation.oid),
            'row_security', relation.relrowsecurity,
            'force_row_security', relation.relforcerowsecurity,
            'columns', COALESCE((SELECT pg_catalog.jsonb_agg(
                pg_catalog.jsonb_build_object(
                    'number', attribute.attnum, 'name', attribute.attname,
                    'type', pg_catalog.format_type(attribute.atttypid, attribute.atttypmod),
                    'not_null', attribute.attnotnull, 'identity', attribute.attidentity,
                    'generated', attribute.attgenerated, 'storage', attribute.attstorage,
                    'compression', attribute.attcompression,
                    'default', pg_catalog.pg_get_expr(attribute_default.adbin, attribute_default.adrelid),
                    'collation', CASE WHEN collation_row.oid IS NULL THEN NULL ELSE
                        pg_catalog.jsonb_build_object(
                            'schema', collation_namespace.nspname, 'name', collation_row.collname,
                            'provider', collation_row.collprovider,
                            'deterministic', collation_row.collisdeterministic,
                            'collate', collation_row.collcollate, 'ctype', collation_row.collctype,
                            'locale', collation_row.colllocale,
                            'icu_rules', collation_row.collicurules,
                            'version', collation_row.collversion) END) ORDER BY attribute.attnum)
                FROM pg_catalog.pg_attribute AS attribute
                LEFT JOIN pg_catalog.pg_attrdef AS attribute_default
                  ON attribute_default.adrelid = attribute.attrelid
                 AND attribute_default.adnum = attribute.attnum
                LEFT JOIN pg_catalog.pg_collation AS collation_row
                  ON collation_row.oid = attribute.attcollation
                LEFT JOIN pg_catalog.pg_namespace AS collation_namespace
                  ON collation_namespace.oid = collation_row.collnamespace
                WHERE attribute.attrelid = relation.oid AND attribute.attnum > 0
                  AND NOT attribute.attisdropped), '[]'::jsonb),
            'constraints', COALESCE((SELECT pg_catalog.jsonb_agg(
                pg_catalog.jsonb_build_object(
                    'name', constraint_row.conname, 'type', constraint_row.contype,
                    'deferrable', constraint_row.condeferrable,
                    'deferred', constraint_row.condeferred,
                    'validated', constraint_row.convalidated,
                    'definition', pg_catalog.pg_get_constraintdef(constraint_row.oid, false))
                ORDER BY constraint_row.conname)
                FROM pg_catalog.pg_constraint AS constraint_row
                WHERE constraint_row.conrelid = relation.oid), '[]'::jsonb),
            'indexes', COALESCE((SELECT pg_catalog.jsonb_agg(
                pg_catalog.jsonb_build_object(
                    'name', index_relation.relname,
                    'definition', pg_catalog.pg_get_indexdef(index_row.indexrelid),
                    'unique', index_row.indisunique, 'primary', index_row.indisprimary,
                    'exclusion', index_row.indisexclusion, 'immediate', index_row.indimmediate,
                    'clustered', index_row.indisclustered, 'valid', index_row.indisvalid,
                    'check_xmin', index_row.indcheckxmin, 'ready', index_row.indisready,
                    'live', index_row.indislive, 'replica_identity', index_row.indisreplident)
                ORDER BY index_relation.relname, pg_catalog.pg_get_indexdef(index_row.indexrelid))
                FROM pg_catalog.pg_index AS index_row
                JOIN pg_catalog.pg_class AS index_relation
                  ON index_relation.oid = index_row.indexrelid
                WHERE index_row.indrelid = relation.oid), '[]'::jsonb),
            'triggers', COALESCE((SELECT pg_catalog.jsonb_agg(
                pg_catalog.jsonb_build_object(
                    'name', trigger_row.tgname, 'enabled', trigger_row.tgenabled,
                    'definition', pg_catalog.pg_get_triggerdef(trigger_row.oid, false),
                    'function_schema', trigger_function_namespace.nspname,
                    'function_name', trigger_function.proname,
                    'function_identity_arguments',
                        pg_catalog.pg_get_function_identity_arguments(trigger_function.oid),
                    'function_owner', pg_catalog.pg_get_userbyid(trigger_function.proowner),
                    'function_definition', pg_catalog.pg_get_functiondef(trigger_function.oid),
                    'function_acl', cutover_authority.catalog_acl_sha256_v1(
                        'function', trigger_function.oid)) ORDER BY trigger_row.tgname)
                FROM pg_catalog.pg_trigger AS trigger_row
                JOIN pg_catalog.pg_proc AS trigger_function
                  ON trigger_function.oid = trigger_row.tgfoid
                JOIN pg_catalog.pg_namespace AS trigger_function_namespace
                  ON trigger_function_namespace.oid = trigger_function.pronamespace
                WHERE trigger_row.tgrelid = relation.oid AND NOT trigger_row.tgisinternal), '[]'::jsonb),
            'rules', COALESCE((SELECT pg_catalog.jsonb_agg(
                pg_catalog.jsonb_build_object('name', rule_row.rulename,
                    'enabled', rule_row.ev_enabled,
                    'definition', pg_catalog.pg_get_ruledef(rule_row.oid, false))
                ORDER BY rule_row.rulename)
                FROM pg_catalog.pg_rewrite AS rule_row
                WHERE rule_row.ev_class = relation.oid AND rule_row.rulename <> '_RETURN'), '[]'::jsonb),
            'policies', COALESCE((SELECT pg_catalog.jsonb_agg(
                pg_catalog.jsonb_build_object('name', policy.polname, 'command', policy.polcmd,
                    'permissive', policy.polpermissive,
                    'roles', COALESCE((SELECT pg_catalog.jsonb_agg(
                        pg_catalog.pg_get_userbyid(role_oid) ORDER BY pg_catalog.pg_get_userbyid(role_oid))
                        FROM pg_catalog.unnest(policy.polroles) AS role_oid), '[]'::jsonb),
                    'qual', pg_catalog.pg_get_expr(policy.polqual, policy.polrelid),
                    'with_check', pg_catalog.pg_get_expr(policy.polwithcheck, policy.polrelid))
                ORDER BY policy.polname) FROM pg_catalog.pg_policy AS policy
                WHERE policy.polrelid = relation.oid), '[]'::jsonb))
            FROM pg_catalog.pg_class AS relation
            JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
            LEFT JOIN pg_catalog.pg_am AS access_method ON access_method.oid = relation.relam
            LEFT JOIN pg_catalog.pg_tablespace AS tablespace ON tablespace.oid = relation.reltablespace
            WHERE p_kind = 'table' AND relation.oid = p_object_oid),
        'routine', (SELECT pg_catalog.jsonb_build_object(
            'schema', namespace.nspname, 'name', procedure_row.proname,
            'identity_arguments', pg_catalog.pg_get_function_identity_arguments(procedure_row.oid),
            'result', pg_catalog.pg_get_function_result(procedure_row.oid),
            'language', language.lanname, 'kind', procedure_row.prokind,
            'volatile', procedure_row.provolatile, 'parallel', procedure_row.proparallel,
            'security_definer', procedure_row.prosecdef, 'strict', procedure_row.proisstrict,
            'leakproof', procedure_row.proleakproof,
            'config', COALESCE((SELECT pg_catalog.jsonb_agg(config_row ORDER BY config_row)
                FROM pg_catalog.unnest(procedure_row.proconfig) AS config_row), '[]'::jsonb),
            'definition', pg_catalog.pg_get_functiondef(procedure_row.oid))
            FROM pg_catalog.pg_proc AS procedure_row
            JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = procedure_row.pronamespace
            JOIN pg_catalog.pg_language AS language ON language.oid = procedure_row.prolang
            WHERE p_kind = 'function' AND procedure_row.oid = p_object_oid)
    )::text, 'UTF8')), 'hex'));
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION cutover_authority.catalog_definition_sha256_v1(text, oid) FROM PUBLIC;
