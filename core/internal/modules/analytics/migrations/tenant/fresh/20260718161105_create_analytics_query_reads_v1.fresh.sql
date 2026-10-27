-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_query_get_v1(p_scope_schema name, p_query_id uuid)
RETURNS TABLE (
    query_id uuid, version_id uuid, version_no bigint,
    root_revision bigint, reference_epoch bigint, version_revision bigint,
    lifecycle_state text, version_state text, name text,
    query_document bytea, parameter_document bytea, output_document bytea,
    manifest_sha256 bytea, cache_policy text,
    created_at timestamptz, updated_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    RETURN QUERY
    SELECT root.id, version_row.id, version_row.version_no,
           root.row_revision, root.reference_epoch, version_row.row_revision,
           root.lifecycle_state, version_row.version_state, version_row.name,
           version_row.query_document, version_row.parameter_document,
           version_row.output_document, version_row.manifest_sha256,
           version_row.cache_policy, root.created_at, root.updated_at
    FROM analytics_saved_queries AS root
    JOIN analytics_saved_query_versions AS version_row
      ON version_row.id = root.current_draft_version_id
     AND version_row.query_id = root.id
    WHERE root.id = p_query_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102',
            MESSAGE = 'analytics query was not found';
    END IF;
END;
$function$;

CREATE FUNCTION analytics_query_get_version_v1(
    p_scope_schema name, p_query_id uuid, p_version_id uuid
)
RETURNS TABLE (
    query_id uuid, version_id uuid, version_no bigint,
    root_revision bigint, reference_epoch bigint, version_revision bigint,
    lifecycle_state text, version_state text, name text,
    query_document bytea, parameter_document bytea, output_document bytea,
    manifest_sha256 bytea, cache_policy text,
    created_at timestamptz, updated_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    RETURN QUERY
    SELECT root.id, version_row.id, version_row.version_no,
           root.row_revision, root.reference_epoch, version_row.row_revision,
           root.lifecycle_state, version_row.version_state, version_row.name,
           version_row.query_document, version_row.parameter_document,
           version_row.output_document, version_row.manifest_sha256,
           version_row.cache_policy, root.created_at, version_row.updated_at
    FROM analytics_saved_queries AS root
    JOIN analytics_saved_query_versions AS version_row
      ON version_row.query_id = root.id
    WHERE root.id = p_query_id AND version_row.id = p_version_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102',
            MESSAGE = 'analytics query version was not found';
    END IF;
END;
$function$;

CREATE FUNCTION analytics_query_list_v1(
    p_scope_schema name, p_limit integer, p_after uuid
)
RETURNS TABLE (
    query_id uuid, version_id uuid, version_no bigint,
    root_revision bigint, reference_epoch bigint, version_revision bigint,
    lifecycle_state text, version_state text, name text,
    query_document bytea, parameter_document bytea, output_document bytea,
    manifest_sha256 bytea, cache_policy text,
    created_at timestamptz, updated_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_limit IS NULL OR p_limit NOT BETWEEN 1 AND 100 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'analytics query list input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    RETURN QUERY
    SELECT root.id, version_row.id, version_row.version_no,
           root.row_revision, root.reference_epoch, version_row.row_revision,
           root.lifecycle_state, version_row.version_state, version_row.name,
           version_row.query_document, version_row.parameter_document,
           version_row.output_document, version_row.manifest_sha256,
           version_row.cache_policy, root.created_at, root.updated_at
    FROM analytics_saved_queries AS root
    JOIN analytics_saved_query_versions AS version_row
      ON version_row.id = root.current_draft_version_id
     AND version_row.query_id = root.id
    WHERE root.lifecycle_state = 'active'
      AND (p_after IS NULL OR root.id > p_after)
    ORDER BY root.id
    LIMIT p_limit;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_query_get_v1(name,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_query_get_version_v1(name,uuid,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_query_list_v1(name,integer,uuid) FROM PUBLIC;

