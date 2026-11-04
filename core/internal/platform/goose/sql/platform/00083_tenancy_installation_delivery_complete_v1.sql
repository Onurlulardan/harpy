-- +goose Up
CREATE FUNCTION platform.tenancy_installation_delivery_complete_v1(
    in_operation_id uuid,
    in_attempt_id uuid,
    in_expected_revision bigint
)
RETURNS TABLE (
    delivery_state text,
    delivery_revision bigint,
    database_now timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        in_expected_revision,
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION platform.tenancy_installation_delivery_complete_v1(uuid, uuid, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_installation_delivery_complete_v1(uuid, uuid, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_installation_delivery_complete_v1(uuid, uuid, bigint);
