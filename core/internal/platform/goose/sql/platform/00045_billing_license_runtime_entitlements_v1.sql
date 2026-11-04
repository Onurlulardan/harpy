-- +goose Up
-- BELIRSIZ: runtime grant rows are not determined by the caller.
CREATE FUNCTION platform.billing_license_runtime_entitlements_v1(
    p_environment_id uuid
)
RETURNS TABLE (
    license_id uuid,
    environment_id uuid,
    snapshot_sequence bigint,
    payload_sha256 text,
    runtime_state text,
    entitlement_key text,
    value_type text,
    boolean_value boolean,
    integer_value bigint,
    unlimited boolean,
    release_channel text,
    database_now timestamptz
)
LANGUAGE sql
STABLE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        NULL::uuid,
        NULL::uuid,
        NULL::bigint,
        NULL::text,
        NULL::text,
        NULL::text,
        NULL::text,
        NULL::boolean,
        NULL::bigint,
        NULL::boolean,
        NULL::text,
        NULL::timestamptz
    WHERE false;
$$;

REVOKE ALL ON FUNCTION platform.billing_license_runtime_entitlements_v1(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_license_runtime_entitlements_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_license_runtime_entitlements_v1(uuid);
