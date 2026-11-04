-- +goose Up
-- BELIRSIZ: pending and fenced are not determined by the caller.
CREATE FUNCTION platform.billing_license_refresh_failure_v1(
    p_operation_id uuid,
    p_lease_token uuid,
    p_failure_kind text
)
RETURNS text
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT ''::text;
$$;

REVOKE ALL ON FUNCTION platform.billing_license_refresh_failure_v1(
    uuid, uuid, text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_license_refresh_failure_v1(
    uuid, uuid, text
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_license_refresh_failure_v1(uuid, uuid, text);
