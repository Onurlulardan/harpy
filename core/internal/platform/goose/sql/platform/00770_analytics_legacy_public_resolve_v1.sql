-- +goose Up
-- BELIRSIZ: no legacy public route is stored.
-- +goose StatementBegin
CREATE FUNCTION platform.analytics_legacy_public_resolve_v1(
    p_tenant_identity text,
    p_method text,
    p_path text,
    p_slug text
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
AS $fn$
    SELECT
        p_tenant_identity,
        p_method,
        p_path,
        p_slug,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        decode('00000000000000000000000000000000', 'hex'),
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        0::bigint,
        NULL::timestamptz,
        NULL::timestamptz,
        pg_catalog.clock_timestamp();
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.analytics_legacy_public_resolve_v1(text, text, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.analytics_legacy_public_resolve_v1(text, text, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.analytics_legacy_public_resolve_v1(text, text, text, text);
