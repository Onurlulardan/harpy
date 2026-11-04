-- +goose Up
-- BELIRSIZ: no public route is stored. public_id is 16 bytes so Scan can copy it.
CREATE FUNCTION platform.analytics_public_resolve_v1(
    p_public_id bytea
)
RETURNS TABLE (
    public_id bytea,
    tenant_id uuid,
    release_id uuid,
    resource_kind text,
    resource_version_id uuid,
    audience text,
    state text,
    generation bigint,
    capability_sha256 bytea,
    capability_expires_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        p_public_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        0::bigint,
        ''::bytea,
        NULL::timestamptz;
$$;

REVOKE ALL ON FUNCTION platform.analytics_public_resolve_v1(bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.analytics_public_resolve_v1(bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.analytics_public_resolve_v1(bytea);
