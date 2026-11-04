-- +goose Up
-- BELIRSIZ: refresh operation state is not determined by the caller.
CREATE FUNCTION platform.billing_platform_license_refresh_read_v1(
    p_environment_id uuid,
    p_operation_id uuid
)
RETURNS TABLE (
    operation_id uuid,
    operation_state text,
    snapshot_sequence bigint,
    snapshot_sha256 text,
    failure_code text,
    created_at timestamptz,
    completed_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        p_operation_id,
        ''::text,
        NULL::bigint,
        NULL::text,
        NULL::text,
        pg_catalog.clock_timestamp(),
        NULL::timestamptz;
$$;

REVOKE ALL ON FUNCTION platform.billing_platform_license_refresh_read_v1(
    uuid, uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_platform_license_refresh_read_v1(
    uuid, uuid
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_platform_license_refresh_read_v1(uuid, uuid);
