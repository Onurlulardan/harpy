-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_installation_delivery_secret_deleted_v1(
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
AS $fn$
    UPDATE platform.tenancy_installation_deliveries AS delivery
    SET delivery_state = 'secret_deleted',
        delivery_revision = delivery.delivery_revision + 1,
        secret_deleted_at = pg_catalog.clock_timestamp(),
        database_now = pg_catalog.clock_timestamp()
    WHERE delivery.operation_id = in_operation_id
      AND delivery.current_attempt_id = in_attempt_id
      AND delivery.delivery_revision = in_expected_revision
      AND delivery.delivery_state = 'acknowledged'
    RETURNING delivery.delivery_state, delivery.delivery_revision, delivery.database_now;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_installation_delivery_secret_deleted_v1(uuid, uuid, bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_installation_delivery_secret_deleted_v1(uuid, uuid, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_installation_delivery_secret_deleted_v1(uuid, uuid, bigint);
