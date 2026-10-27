-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_validate_submission_step_v1(
    p_plan_sha256 text,
    p_resource_kind text,
    p_document bytea,
    p_expected_ordinal integer,
    p_allowed_operations jsonb
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_document jsonb;
    v_step jsonb;
    v_operation text;
    v_executor text;
    v_transaction_class text;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR p_plan_sha256 !~ '^[0-9a-f]{64}$'
       OR p_resource_kind NOT IN ('form_storage', 'form_purge')
       OR p_expected_ordinal NOT BETWEEN 0 AND 1999
       OR p_document IS NULL
       OR pg_catalog.octet_length(p_document) NOT BETWEEN 2 AND 1048576
       OR pg_catalog.encode(pg_catalog.sha256(p_document), 'hex') IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'schema submission step input is invalid';
    END IF;
    BEGIN
        v_document := pg_catalog.convert_from(p_document, 'UTF8')::jsonb;
    EXCEPTION
        WHEN character_not_in_repertoire OR invalid_text_representation THEN
            RAISE EXCEPTION USING ERRCODE = '22023',
                MESSAGE = 'schema submission step document is invalid';
    END;
    IF pg_catalog.jsonb_typeof(v_document) <> 'object'
       OR (SELECT pg_catalog.count(*) FROM
            pg_catalog.jsonb_object_keys(v_document)) <> 2
       OR NOT (v_document ? 'plan_sha256' AND v_document ? 'step')
       OR v_document->>'plan_sha256' IS DISTINCT FROM p_plan_sha256
       OR pg_catalog.jsonb_typeof(v_document->'step') <> 'object' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'schema submission step envelope differs';
    END IF;
    v_step := v_document->'step';
    IF (SELECT pg_catalog.count(*) FROM
            pg_catalog.jsonb_object_keys(v_step)) <> 6
       OR NOT (v_step ?& ARRAY[
            'ordinal', 'operation_id', 'transaction_class', 'executor_kind',
            'identifier_mapping', 'arguments'
       ])
       OR (v_step->>'ordinal')::integer IS DISTINCT FROM p_expected_ordinal
       OR pg_catalog.jsonb_typeof(v_step->'identifier_mapping') <> 'object'
       OR v_step->'identifier_mapping' = '{}'::jsonb
       OR pg_catalog.jsonb_typeof(v_step->'arguments') <> 'object'
       OR pg_catalog.octet_length((v_step->'identifier_mapping')::text)
            NOT BETWEEN 2 AND 65536
       OR pg_catalog.octet_length((v_step->'arguments')::text)
            NOT BETWEEN 2 AND 262144
       OR v_step::text ~*
            '"(sql|raw_sql|function_body|routine_body|client_identifier)"[[:space:]]*:'
       THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'schema submission step shape differs';
    END IF;
    v_operation := v_step->>'operation_id';
    v_executor := v_step->>'executor_kind';
    v_transaction_class := v_step->>'transaction_class';
    IF NOT EXISTS (
        SELECT 1
        FROM pg_catalog.jsonb_array_elements_text(p_allowed_operations) AS allowed(value)
        WHERE allowed.value = v_operation
    ) OR v_operation NOT IN (
        'activate_generation', 'add_constraint', 'add_field_column',
        'add_relation_table', 'backfill_field_batch', 'cleanup_invalid_index',
        'create_form_table', 'create_index', 'create_index_concurrently',
        'disable_old_writer', 'drop_constraint', 'drop_field_column',
        'drop_form_table', 'drop_index', 'drop_index_concurrently',
        'drop_relation_table', 'enable_dual_write', 'retain_generation',
        'validate_constraint', 'verify_storage'
    ) OR (v_operation = 'drop_form_table' AND p_resource_kind <> 'form_purge') THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'schema submission step operation is invalid';
    END IF;
    IF (
        v_operation IN (
            'cleanup_invalid_index', 'create_index_concurrently',
            'drop_index_concurrently'
        ) AND (
            v_executor <> 'migrator'
            OR v_transaction_class <> 'concurrent_index'
        )
    ) OR (
        v_operation IN (
            'activate_generation', 'disable_old_writer',
            'enable_dual_write', 'retain_generation'
        ) AND (
            v_executor <> 'forms_runtime'
            OR v_transaction_class <> 'transactional'
        )
    ) OR (
        v_operation NOT IN (
            'cleanup_invalid_index', 'create_index_concurrently',
            'drop_index_concurrently', 'activate_generation',
            'disable_old_writer', 'enable_dual_write', 'retain_generation'
        ) AND (
            v_executor <> 'migrator'
            OR v_transaction_class <> 'transactional'
        )
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'schema submission execution contract differs';
    END IF;
    RETURN v_step;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
platform.platform_migrator_validate_submission_step_v1(
    text, text, bytea, integer, jsonb
) FROM PUBLIC;
