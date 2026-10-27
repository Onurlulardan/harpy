-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_assert_schema_submission_v1(
    p_scope_schema name,
    p_tenant_id uuid,
    p_job_id uuid,
    p_resource_kind text,
    p_resource_id uuid,
    p_resource_revision bigint,
    p_plan_contract_version text,
    p_compiler_version text,
    p_postgresql_major integer,
    p_canonical_plan_bytes bytea,
    p_plan_sha256 text,
    p_canonical_steps jsonb
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_plan jsonb;
    v_allowed jsonb;
    v_operation text;
    v_previous text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR p_scope_schema IS NULL
       OR p_scope_schema::text !~ '^[a-z][a-z0-9_]{0,62}$'
       OR pg_catalog.uuid_extract_version(p_tenant_id) IS DISTINCT FROM 7
       OR pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7
       OR pg_catalog.uuid_extract_version(p_resource_id) IS DISTINCT FROM 7
       OR p_resource_kind NOT IN ('form_storage', 'form_purge')
       OR p_resource_revision NOT BETWEEN 1 AND 9007199254740991
       OR p_plan_contract_version <> 'forms-schema-plan.v1'
       OR p_compiler_version <> '1.0.0'
       OR p_postgresql_major <> 18
       OR p_plan_sha256 !~ '^[0-9a-f]{64}$'
       OR p_canonical_plan_bytes IS NULL
       OR pg_catalog.octet_length(p_canonical_plan_bytes) NOT BETWEEN 2 AND 1048576
       OR pg_catalog.jsonb_typeof(p_canonical_steps) IS DISTINCT FROM 'array' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'schema submission identity is invalid';
    END IF;
    IF pg_catalog.jsonb_array_length(p_canonical_steps) NOT BETWEEN 1 AND 2000
       OR pg_catalog.octet_length(p_canonical_steps::text) > 4194304
       OR EXISTS (
            SELECT 1
            FROM pg_catalog.jsonb_array_elements(p_canonical_steps) AS encoded(value)
            WHERE pg_catalog.jsonb_typeof(encoded.value) <> 'string'
               OR encoded.value #>> '{}' !~ '^([0-9a-f]{2})+$'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'schema submission step transport is invalid';
    END IF;
    IF NOT EXISTS (
        SELECT 1
        FROM platform.tenants AS tenant
        WHERE tenant.id = p_tenant_id
          AND tenant.schema_name = p_scope_schema::text
          AND tenant.deleted_at IS NULL
          AND tenant.status IN ('ACTIVE', 'TRIAL')
        FOR SHARE
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'schema submission TenantScope is invalid';
    END IF;
    PERFORM platform.billing_resolve_current_tenant_authority_v1(p_tenant_id);
    IF pg_catalog.encode(
        pg_catalog.sha256(p_canonical_plan_bytes), 'hex'
    ) IS DISTINCT FROM p_plan_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'schema submission plan digest differs';
    END IF;
    BEGIN
        v_plan := pg_catalog.convert_from(p_canonical_plan_bytes, 'UTF8')::jsonb;
    EXCEPTION
        WHEN character_not_in_repertoire OR invalid_text_representation THEN
            RAISE EXCEPTION USING ERRCODE = '22023',
                MESSAGE = 'schema submission plan document is invalid';
    END;
    IF pg_catalog.jsonb_typeof(v_plan) <> 'object'
       OR (SELECT pg_catalog.count(*) FROM
            pg_catalog.jsonb_object_keys(v_plan)) <> 7
       OR NOT (v_plan ?& ARRAY[
            'contract_version', 'tenant_id', 'owner', 'resource', 'compiler',
            'allowed_operations', 'steps'
       ])
       OR v_plan->>'contract_version' IS DISTINCT FROM p_plan_contract_version
       OR v_plan->>'tenant_id' IS DISTINCT FROM p_tenant_id::text
       OR v_plan->>'owner' IS DISTINCT FROM 'forms'
       OR pg_catalog.jsonb_typeof(v_plan->'resource') <> 'object'
       OR pg_catalog.jsonb_typeof(v_plan->'compiler') <> 'object'
       OR pg_catalog.jsonb_typeof(v_plan->'allowed_operations') <> 'array'
       OR pg_catalog.jsonb_typeof(v_plan->'steps') <> 'array'
       OR v_plan::text ~*
            '"(sql|raw_sql|function_body|routine_body|client_identifier)"[[:space:]]*:'
       THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'schema submission plan shape differs';
    END IF;
    IF (SELECT pg_catalog.count(*) FROM
            pg_catalog.jsonb_object_keys(v_plan->'resource')) <> 3
       OR NOT ((v_plan->'resource') ?& ARRAY['kind', 'id', 'revision'])
       OR v_plan->'resource'->>'kind' IS DISTINCT FROM p_resource_kind
       OR v_plan->'resource'->>'id' IS DISTINCT FROM p_resource_id::text
       OR (v_plan->'resource'->>'revision')::bigint
            IS DISTINCT FROM p_resource_revision
       OR (SELECT pg_catalog.count(*) FROM
            pg_catalog.jsonb_object_keys(v_plan->'compiler')) <> 2
       OR NOT ((v_plan->'compiler') ?& ARRAY['version', 'postgresql_major'])
       OR v_plan->'compiler'->>'version' IS DISTINCT FROM p_compiler_version
       OR (v_plan->'compiler'->>'postgresql_major')::integer
            IS DISTINCT FROM p_postgresql_major
       OR pg_catalog.jsonb_array_length(v_plan->'steps')
            IS DISTINCT FROM pg_catalog.jsonb_array_length(p_canonical_steps)
       OR v_plan->'steps'->0->>'executor_kind' IS DISTINCT FROM 'migrator' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'schema submission resource contract differs';
    END IF;
    v_allowed := v_plan->'allowed_operations';
    IF pg_catalog.jsonb_array_length(v_allowed) NOT BETWEEN 1 AND 20 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'schema submission operation allowlist is invalid';
    END IF;
    FOR v_operation IN
        SELECT allowed.value
        FROM pg_catalog.jsonb_array_elements_text(v_allowed)
            WITH ORDINALITY AS allowed(value, ordinal)
        ORDER BY allowed.ordinal
    LOOP
        IF v_operation NOT IN (
            'activate_generation', 'add_constraint', 'add_field_column',
            'add_relation_table', 'backfill_field_batch', 'cleanup_invalid_index',
            'create_form_table', 'create_index', 'create_index_concurrently',
            'disable_old_writer', 'drop_constraint', 'drop_field_column',
            'drop_form_table', 'drop_index', 'drop_index_concurrently',
            'drop_relation_table', 'enable_dual_write', 'retain_generation',
            'validate_constraint', 'verify_storage'
        ) OR (v_previous IS NOT NULL AND v_previous >= v_operation) THEN
            RAISE EXCEPTION USING ERRCODE = '22023',
                MESSAGE = 'schema submission operation allowlist differs';
        END IF;
        v_previous := v_operation;
    END LOOP;
    RETURN v_plan;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
platform.platform_migrator_assert_schema_submission_v1(
    name, uuid, uuid, text, uuid, bigint, text, text, integer, bytea, text, jsonb
) FROM PUBLIC;
