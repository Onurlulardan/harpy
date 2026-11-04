-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION analytics_query_latest_published_v1(
    arg_1 name,
    arg_2 uuid
)
RETURNS TABLE (
    query_id uuid,
    version_id uuid,
    version_no bigint,
    root_revision bigint,
    reference_epoch bigint,
    version_revision bigint,
    lifecycle_state text,
    version_state text,
    "name" text,
    query_document bytea,
    parameter_document bytea,
    output_document bytea,
    manifest_sha256 bytea,
    cache_policy text,
    created_at timestamptz,
    updated_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        ''::text,
        ''::text,
        ''::text,
        NULL::bytea,
        NULL::bytea,
        NULL::bytea,
        NULL::bytea,
        ''::text,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION analytics_query_latest_published_v1(name, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_query_latest_published_v1(name, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_query_latest_published_v1(name, uuid);
