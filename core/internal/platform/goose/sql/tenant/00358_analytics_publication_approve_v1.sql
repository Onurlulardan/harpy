-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION analytics_publication_approve_v1(
    arg_1 name,
    arg_2 uuid,
    arg_3 uuid,
    arg_4 text,
    arg_5 bytea,
    arg_6 uuid,
    arg_7 bigint,
    arg_8 timestamptz,
    arg_9 timestamptz
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
    updated_at timestamptz,
    replayed boolean
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
        pg_catalog.clock_timestamp(),
        false;
$$;

REVOKE ALL ON FUNCTION analytics_publication_approve_v1(name, uuid, uuid, text, bytea, uuid, bigint, timestamptz, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_publication_approve_v1(name, uuid, uuid, text, bytea, uuid, bigint, timestamptz, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_publication_approve_v1(name, uuid, uuid, text, bytea, uuid, bigint, timestamptz, timestamptz);
