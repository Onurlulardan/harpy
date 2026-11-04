-- +goose Up
CREATE FUNCTION platform.tenancy_installation_delivery_unknown_v1(
    in_operation_id uuid,
    in_attempt_id uuid,
    in_expected_revision bigint,
    in_adapter_id text,
    in_claim_token_sha256 text,
    in_cause text
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
    UPDATE platform.tenancy_installation_deliveries AS delivery
    SET delivery_state = 'outcome_unknown',
        delivery_revision = delivery.delivery_revision + 1,
        recovery_cause = in_cause,
        database_now = pg_catalog.clock_timestamp()
    WHERE delivery.operation_id = in_operation_id
      AND delivery.current_attempt_id = in_attempt_id
      AND delivery.delivery_revision = in_expected_revision
      AND delivery.claim_adapter_id = in_adapter_id
      AND delivery.delivery_state = 'claimed'
    RETURNING delivery.delivery_state, delivery.delivery_revision, delivery.database_now;
$$;

REVOKE ALL ON FUNCTION platform.tenancy_installation_delivery_unknown_v1(uuid, uuid, bigint, text, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_installation_delivery_unknown_v1(uuid, uuid, bigint, text, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_installation_delivery_unknown_v1(uuid, uuid, bigint, text, text, text);
