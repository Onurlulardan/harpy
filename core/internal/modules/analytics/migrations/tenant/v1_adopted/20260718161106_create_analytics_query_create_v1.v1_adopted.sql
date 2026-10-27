-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_query_create_v1(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid,
    p_idempotency_key text, p_request_sha256 bytea,
    p_query_id uuid, p_version_id uuid, p_name text,
    p_query_document bytea, p_parameter_document bytea,
    p_output_document bytea, p_occurred_at timestamptz
)
RETURNS TABLE (
    query_id uuid, version_id uuid, version_no bigint,
    root_revision bigint, reference_epoch bigint, version_revision bigint,
    lifecycle_state text, version_state text, name text,
    query_document bytea, parameter_document bytea, output_document bytea,
    manifest_sha256 bytea, cache_policy text,
    created_at timestamptz, updated_at timestamptz, replayed boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_existing analytics_operations%ROWTYPE;
    v_manifest bytea;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_query_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_version_id) IS DISTINCT FROM 7 OR
       p_actor_id IS NULL OR p_actor_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       p_idempotency_key IS NULL OR
       pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       p_idempotency_key OPERATOR(pg_catalog.!~) '^[A-Za-z0-9_.:-]+$' OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_name) NOT BETWEEN 1 AND 100 OR
       pg_catalog.octet_length(p_query_document) NOT BETWEEN 2 AND 65536 OR
       pg_catalog.octet_length(p_parameter_document) NOT BETWEEN 2 AND 65536 OR
       pg_catalog.octet_length(p_output_document) NOT BETWEEN 2 AND 65536 OR
       NOT pg_catalog.pg_input_is_valid(pg_catalog.convert_from(p_query_document, 'UTF8'), 'jsonb') OR
       NOT pg_catalog.pg_input_is_valid(pg_catalog.convert_from(p_parameter_document, 'UTF8'), 'jsonb') OR
       NOT pg_catalog.pg_input_is_valid(pg_catalog.convert_from(p_output_document, 'UTF8'), 'jsonb') OR
       p_occurred_at IS NULL OR NOT pg_catalog.isfinite(p_occurred_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'analytics query create input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:query:create:' || p_actor_id::text || ':' || p_idempotency_key, 6100));
    SELECT operation.* INTO v_existing
    FROM analytics_operations AS operation
    WHERE operation.actor_id = p_actor_id
      AND operation.operation_kind = 'query_create'
      AND operation.idempotency_key = p_idempotency_key
    FOR UPDATE;
    IF FOUND THEN
        IF v_existing.request_sha256 IS DISTINCT FROM p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = 'P6103',
                MESSAGE = 'analytics query idempotency key was reused';
        END IF;
        RETURN QUERY
        SELECT result.*, true
        FROM analytics_query_get_v1(p_scope_schema, v_existing.resource_id) AS result;
        RETURN;
    END IF;
    IF EXISTS (SELECT 1 FROM analytics_saved_queries WHERE id = p_query_id) OR
       EXISTS (SELECT 1 FROM analytics_saved_query_versions WHERE id = p_version_id) OR
       EXISTS (SELECT 1 FROM analytics_operations WHERE operation_id = p_operation_id) THEN
        RAISE EXCEPTION USING ERRCODE = 'P6104',
            MESSAGE = 'analytics query stable identity already exists';
    END IF;
    v_manifest := pg_catalog.sha256(
        p_query_document || p_parameter_document || p_output_document);
    SET CONSTRAINTS analytics_saved_queries_current_draft_fk DEFERRED;
    INSERT INTO analytics_saved_queries (
        id, lifecycle_state, row_revision, reference_epoch,
        current_draft_version_id, created_by, updated_by, created_at, updated_at
    ) VALUES (
        p_query_id, 'active', 1, 1, p_version_id,
        p_actor_id, p_actor_id, p_occurred_at, p_occurred_at
    );
    INSERT INTO analytics_saved_query_versions (
        id, query_id, version_no, version_state, row_revision, name,
        query_document, query_sha256, parameter_document, parameter_sha256,
        output_document, output_sha256, manifest_sha256, cache_policy,
        created_by, updated_by, created_at, updated_at
    ) VALUES (
        p_version_id, p_query_id, 1, 'draft', 1, p_name,
        p_query_document, pg_catalog.sha256(p_query_document),
        p_parameter_document, pg_catalog.sha256(p_parameter_document),
        p_output_document, pg_catalog.sha256(p_output_document),
        v_manifest, 'none', p_actor_id, p_actor_id, p_occurred_at, p_occurred_at
    );
    INSERT INTO analytics_operations (
        operation_id, actor_id, operation_kind, idempotency_key,
        request_sha256, resource_kind, resource_id, result_revision,
        result_sha256, occurred_at
    ) VALUES (
        p_operation_id, p_actor_id, 'query_create', p_idempotency_key,
        p_request_sha256, 'query', p_query_id, 1, v_manifest, p_occurred_at
    );
    RETURN QUERY
    SELECT result.*, false
    FROM analytics_query_get_v1(p_scope_schema, p_query_id) AS result;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_query_create_v1(
    name,uuid,uuid,text,bytea,uuid,uuid,text,bytea,bytea,bytea,timestamptz
) FROM PUBLIC;

