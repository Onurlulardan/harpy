-- +goose Up
-- BELIRSIZ: target selection, readiness, and lease are not determined by the caller.
CREATE FUNCTION platform.billing_license_refresh_target_acquire_v1()
RETURNS TABLE (
    target_found boolean,
    ready boolean,
    operation_id uuid,
    license_id uuid,
    licensee_organization_id uuid,
    environment_id uuid,
    delivery_kind text,
    target_revision bigint,
    retry_at timestamptz,
    database_now timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        false,
        false,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        0::bigint,
        clock.database_now,
        clock.database_now
    FROM (SELECT pg_catalog.clock_timestamp() AS database_now) AS clock;
$$;

REVOKE ALL ON FUNCTION platform.billing_license_refresh_target_acquire_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_license_refresh_target_acquire_v1() TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_license_refresh_target_acquire_v1();
