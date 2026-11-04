-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_installation_delivery_ack_v1(
    in_operation_id uuid,
    in_attempt_id uuid,
    in_expected_revision bigint,
    in_adapter_id text,
    in_claim_token_sha256 text
)
RETURNS TABLE (
    operation_id uuid,
    tenant_id uuid,
    first_owner_user_id uuid,
    current_attempt_id uuid,
    current_credential_id uuid,
    current_credential_version bigint,
    delivery_state text,
    delivery_revision bigint,
    database_now timestamptz,
    claim_adapter_id text,
    claim_lease_until timestamptz,
    acknowledged_at timestamptz,
    secret_deleted_at timestamptz,
    completed_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    UPDATE platform.tenancy_installation_deliveries AS delivery
    SET delivery_state = 'acknowledged',
        delivery_revision = delivery.delivery_revision + 1,
        acknowledged_at = pg_catalog.clock_timestamp(),
        database_now = pg_catalog.clock_timestamp()
    WHERE delivery.operation_id = in_operation_id
      AND delivery.current_attempt_id = in_attempt_id
      AND delivery.delivery_revision = in_expected_revision
      AND delivery.claim_adapter_id = in_adapter_id
      AND delivery.delivery_state IN ('claimed', 'outcome_unknown')
    RETURNING delivery.operation_id, delivery.tenant_id, delivery.first_owner_user_id,
              delivery.current_attempt_id, delivery.current_credential_id,
              delivery.current_credential_version, delivery.delivery_state,
              delivery.delivery_revision, delivery.database_now, delivery.claim_adapter_id,
              delivery.claim_lease_until, delivery.acknowledged_at,
              delivery.secret_deleted_at, delivery.completed_at;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_installation_delivery_ack_v1(uuid, uuid, bigint, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_installation_delivery_ack_v1(uuid, uuid, bigint, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_installation_delivery_ack_v1(uuid, uuid, bigint, text, text);
