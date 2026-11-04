-- +goose Up
CREATE FUNCTION platform.billing_license_clock_read_v1()
RETURNS TABLE (
    database_now timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION platform.billing_license_clock_read_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_license_clock_read_v1() TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_license_clock_read_v1();
