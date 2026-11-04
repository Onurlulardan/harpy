-- +goose Up
-- BELIRSIZ: the hit is not stored.
CREATE FUNCTION platform.analytics_legacy_public_record_hit_v1(
    p_tenant_identity text,
    p_method text,
    p_path text,
    p_slug text,
    p_tenant_id uuid,
    p_legacy_endpoint_id uuid,
    p_public_id bytea,
    p_release_id uuid,
    p_at timestamptz
)
RETURNS TABLE (
    tenant_identity text,
    method text,
    path text,
    slug text,
    tenant_id uuid,
    legacy_endpoint_id uuid,
    public_id bytea,
    release_id uuid,
    state text,
    hit_count bigint,
    last_hit_at timestamptz,
    consumers_acked_at timestamptz,
    remove_after timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        p_tenant_identity,
        p_method,
        p_path,
        p_slug,
        p_tenant_id,
        p_legacy_endpoint_id,
        p_public_id,
        p_release_id,
        ''::text,
        0::bigint,
        p_at,
        NULL::timestamptz,
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION platform.analytics_legacy_public_record_hit_v1(text, text, text, text, uuid, uuid, bytea, uuid, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.analytics_legacy_public_record_hit_v1(text, text, text, text, uuid, uuid, bytea, uuid, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.analytics_legacy_public_record_hit_v1(text, text, text, text, uuid, uuid, bytea, uuid, timestamptz);
