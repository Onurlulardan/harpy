-- +goose Up
-- BELIRSIZ: refresh request identity and replay are not determined by the caller.
CREATE FUNCTION platform.billing_platform_license_refresh_request_v1(
    p_environment_id uuid,
    p_tenant_id uuid,
    p_actor_id uuid,
    p_idempotency_key text,
    p_request_sha256 text
)
RETURNS TABLE (
    operation_id uuid,
    replayed boolean,
    created_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        false,
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION platform.billing_platform_license_refresh_request_v1(
    uuid, uuid, uuid, text, text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_platform_license_refresh_request_v1(
    uuid, uuid, uuid, text, text
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_platform_license_refresh_request_v1(
    uuid, uuid, uuid, text, text
);
