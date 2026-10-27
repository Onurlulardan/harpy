-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_query_latest_published_v1(
    p_scope_schema name, p_query_id uuid
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
    JOIN LATERAL (
        SELECT candidate.*
        FROM analytics_saved_query_versions AS candidate
        WHERE candidate.query_id = root.id
          AND candidate.version_state IN ('published', 'needs_revalidation')
        ORDER BY candidate.version_no DESC
        LIMIT 1
    ) AS version_row ON true
    WHERE root.id = p_query_id AND root.lifecycle_state = 'active';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102',
            MESSAGE = 'published analytics query was not found';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_query_latest_published_v1(name,uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_query_latest_published_v1(name,uuid)
    TO harpy_runtime;
