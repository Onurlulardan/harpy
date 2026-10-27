-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_query_publish_v1(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid,
    p_idempotency_key text, p_request_sha256 bytea,
    p_query_id uuid, p_draft_version_id uuid, p_next_draft_version_id uuid,
    p_expected_root_revision bigint, p_expected_draft_revision bigint,
    p_query_document bytea, p_parameter_document bytea, p_output_document bytea,
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
BEGIN
    RETURN QUERY SELECT * FROM analytics_query_publish_core_v1(
        p_scope_schema, 'query_publish', p_operation_id, p_actor_id,
        p_idempotency_key, p_request_sha256, p_query_id, p_draft_version_id,
        p_next_draft_version_id, p_expected_root_revision,
        p_expected_draft_revision, p_query_document, p_parameter_document,
        p_output_document, p_dependencies, p_manifest_sha256,
        p_cache_policy, p_occurred_at
    );
END;
$function$;

CREATE FUNCTION analytics_query_revalidate_v1(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid,
    p_idempotency_key text, p_request_sha256 bytea,
    p_query_id uuid, p_draft_version_id uuid, p_next_draft_version_id uuid,
    p_expected_root_revision bigint, p_expected_draft_revision bigint,
    p_query_document bytea, p_parameter_document bytea, p_output_document bytea,
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
BEGIN
    RETURN QUERY SELECT * FROM analytics_query_publish_core_v1(
        p_scope_schema, 'query_revalidate', p_operation_id, p_actor_id,
        p_idempotency_key, p_request_sha256, p_query_id, p_draft_version_id,
        p_next_draft_version_id, p_expected_root_revision,
        p_expected_draft_revision, p_query_document, p_parameter_document,
        p_output_document, p_dependencies, p_manifest_sha256,
        p_cache_policy, p_occurred_at
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_query_publish_v1(
    name,uuid,uuid,text,bytea,uuid,uuid,uuid,bigint,bigint,
    bytea,bytea,bytea,jsonb,bytea,text,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_query_revalidate_v1(
    name,uuid,uuid,text,bytea,uuid,uuid,uuid,bigint,bigint,
    bytea,bytea,bytea,jsonb,bytea,text,timestamptz
) FROM PUBLIC;
