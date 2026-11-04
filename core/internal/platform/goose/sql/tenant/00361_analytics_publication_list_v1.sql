-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION analytics_publication_list_v1(
    arg_1 name,
    arg_2 integer,
    arg_3 uuid
)
RETURNS TABLE (
    release_id uuid,
    resource_kind text,
    resource_version_id uuid,
    audience text,
    lifecycle_state text,
    cache_policy text,
    manifest_sha256 bytea,
    requester_id uuid,
    approver_id uuid,
    row_revision bigint,
    generation bigint,
    public_id bytea,
    capability_expires_at timestamptz,
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
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        ''::text,
        NULL::bytea,
        '00000000-0000-0000-0000-000000000000'::uuid,
        NULL::uuid,
        0::bigint,
        0::bigint,
        NULL::bytea,
        NULL::timestamptz,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION analytics_publication_list_v1(name, integer, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_publication_list_v1(name, integer, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_publication_list_v1(name, integer, uuid);
