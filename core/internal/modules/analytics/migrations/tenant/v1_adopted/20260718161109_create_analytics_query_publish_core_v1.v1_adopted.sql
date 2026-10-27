-- +goose Up
ALTER TABLE analytics_operations
    ADD COLUMN result_version_id uuid,
    ADD CONSTRAINT analytics_operations_result_version_check CHECK (
        result_version_id IS NULL OR pg_catalog.uuid_extract_version(result_version_id) = 7
    );

-- +goose StatementBegin
CREATE FUNCTION analytics_query_publish_core_v1(
    p_scope_schema name, p_operation_kind text,
    p_operation_id uuid, p_actor_id uuid, p_idempotency_key text,
    p_request_sha256 bytea, p_query_id uuid, p_draft_version_id uuid,
    p_next_draft_version_id uuid, p_expected_root_revision bigint,
    p_expected_draft_revision bigint, p_query_document bytea,
    p_parameter_document bytea, p_output_document bytea,
    p_dependencies jsonb, p_manifest_sha256 bytea, p_cache_policy text,
    p_occurred_at timestamptz
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
    v_version_no bigint;
    v_name text;
    v_count integer;
    v_draft_manifest bytea;
BEGIN
    IF p_operation_kind NOT IN ('query_publish', 'query_revalidate') OR
       pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_query_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_draft_version_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_next_draft_version_id) IS DISTINCT FROM 7 OR
       p_actor_id IS NULL OR p_actor_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_expected_root_revision < 1 OR p_expected_draft_revision < 1 OR
       pg_catalog.octet_length(p_query_document) NOT BETWEEN 2 AND 65536 OR
       pg_catalog.octet_length(p_parameter_document) NOT BETWEEN 2 AND 65536 OR
       pg_catalog.octet_length(p_output_document) NOT BETWEEN 2 AND 65536 OR
       pg_catalog.jsonb_typeof(p_dependencies) IS DISTINCT FROM 'array' OR
       pg_catalog.jsonb_array_length(p_dependencies) NOT BETWEEN 1 AND 32 OR
       pg_catalog.octet_length(p_manifest_sha256) IS DISTINCT FROM 32 OR
       p_cache_policy NOT IN ('none', 'revision_bound') OR
       p_occurred_at IS NULL OR NOT pg_catalog.isfinite(p_occurred_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'analytics query publish input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:query:publish:' || p_actor_id::text || ':' || p_idempotency_key, 6100));
    SELECT operation.* INTO v_existing
    FROM analytics_operations AS operation
    WHERE operation.actor_id = p_actor_id
      AND operation.operation_kind = p_operation_kind
      AND operation.idempotency_key = p_idempotency_key
    FOR UPDATE;
    IF FOUND THEN
        IF v_existing.request_sha256 IS DISTINCT FROM p_request_sha256 OR
           v_existing.result_version_id IS NULL THEN
            RAISE EXCEPTION USING ERRCODE = 'P6103',
                MESSAGE = 'analytics query publish replay differs';
        END IF;
        RETURN QUERY SELECT result.*, true
        FROM analytics_query_get_version_v1(
            p_scope_schema, v_existing.resource_id, v_existing.result_version_id
        ) AS result;
        RETURN;
    END IF;
    SELECT version_row.version_no, version_row.name
    INTO v_version_no, v_name
    FROM analytics_saved_queries AS root
    JOIN analytics_saved_query_versions AS version_row
      ON version_row.id = root.current_draft_version_id
     AND version_row.query_id = root.id
    WHERE root.id = p_query_id AND root.lifecycle_state = 'active'
      AND root.row_revision = p_expected_root_revision
      AND version_row.id = p_draft_version_id
      AND version_row.version_state = 'draft'
      AND version_row.row_revision = p_expected_draft_revision
    FOR UPDATE OF root, version_row;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'analytics query publish revision changed';
    END IF;
    IF EXISTS (SELECT 1 FROM analytics_saved_query_versions
               WHERE id = p_next_draft_version_id) THEN
        RAISE EXCEPTION USING ERRCODE = 'P6104',
            MESSAGE = 'analytics query version identity already exists';
    END IF;
    INSERT INTO analytics_query_dependencies (
        query_version_id, ordinal, owner_domain, resource_kind,
        resource_id, resource_version_id, owner_generation,
        catalog_version, query_contract_version, policy_version,
        data_revision_capability, dependency_sha256, created_at
    )
    SELECT p_draft_version_id, (item.ordinality - 1)::smallint,
        item.value->>'owner_domain', item.value->>'resource_kind',
        (item.value->>'resource_id')::uuid,
        (item.value->>'resource_version_id')::uuid,
        (item.value->>'owner_generation')::bigint,
        (item.value->>'catalog_version')::bigint,
        (item.value->>'query_contract_version')::bigint,
        (item.value->>'policy_version')::bigint,
        COALESCE(item.value->>'data_revision_capability', ''),
        pg_catalog.decode(item.value->>'dependency_sha256', 'hex'),
        p_occurred_at
    FROM pg_catalog.jsonb_array_elements(p_dependencies)
         WITH ORDINALITY AS item(value, ordinality);
    GET DIAGNOSTICS v_count = ROW_COUNT;
    IF v_count IS DISTINCT FROM pg_catalog.jsonb_array_length(p_dependencies) OR
       (p_cache_policy = 'revision_bound' AND EXISTS (
           SELECT 1 FROM analytics_query_dependencies
           WHERE query_version_id = p_draft_version_id
             AND data_revision_capability = ''
       )) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'analytics query dependency manifest is invalid';
    END IF;
    UPDATE analytics_saved_query_versions
    SET version_state = 'published', row_revision = row_revision + 1,
        query_document = p_query_document, query_sha256 = pg_catalog.sha256(p_query_document),
        parameter_document = p_parameter_document,
        parameter_sha256 = pg_catalog.sha256(p_parameter_document),
        output_document = p_output_document, output_sha256 = pg_catalog.sha256(p_output_document),
        manifest_sha256 = p_manifest_sha256, cache_policy = p_cache_policy,
        updated_by = p_actor_id, updated_at = p_occurred_at,
        published_by = p_actor_id, published_at = p_occurred_at
    WHERE id = p_draft_version_id;
    v_draft_manifest := pg_catalog.sha256(
        p_query_document || p_parameter_document || p_output_document);
    SET CONSTRAINTS analytics_saved_queries_current_draft_fk DEFERRED;
    INSERT INTO analytics_saved_query_versions (
        id, query_id, version_no, version_state, row_revision, name,
        query_document, query_sha256, parameter_document, parameter_sha256,
        output_document, output_sha256, manifest_sha256, cache_policy,
        created_by, updated_by, created_at, updated_at
    ) VALUES (
        p_next_draft_version_id, p_query_id, v_version_no + 1, 'draft', 1, v_name,
        p_query_document, pg_catalog.sha256(p_query_document),
        p_parameter_document, pg_catalog.sha256(p_parameter_document),
        p_output_document, pg_catalog.sha256(p_output_document),
        v_draft_manifest, 'none', p_actor_id, p_actor_id, p_occurred_at, p_occurred_at
    );
    UPDATE analytics_saved_queries
    SET current_draft_version_id = p_next_draft_version_id,
        row_revision = row_revision + 1, updated_by = p_actor_id, updated_at = p_occurred_at
    WHERE id = p_query_id;
    INSERT INTO analytics_operations (
        operation_id, actor_id, operation_kind, idempotency_key,
        request_sha256, resource_kind, resource_id, result_revision,
        result_sha256, occurred_at, result_version_id
    ) VALUES (
        p_operation_id, p_actor_id, p_operation_kind, p_idempotency_key,
        p_request_sha256, 'query', p_query_id, p_expected_root_revision + 1,
        p_manifest_sha256, p_occurred_at, p_draft_version_id
    );
    RETURN QUERY SELECT result.*, false
    FROM analytics_query_get_version_v1(
        p_scope_schema, p_query_id, p_draft_version_id
    ) AS result;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_query_publish_core_v1(
    name,text,uuid,uuid,text,bytea,uuid,uuid,uuid,bigint,bigint,
    bytea,bytea,bytea,jsonb,bytea,text,timestamptz
) FROM PUBLIC;

