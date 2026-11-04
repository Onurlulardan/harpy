-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_installation_delivery_record_v1(
    in_operation_id uuid,
    in_current_attempt_id uuid,
    in_current_credential_id uuid,
    in_current_credential_version bigint,
    in_next_attempt_id uuid,
    in_next_credential_id uuid,
    in_next_credential_version bigint,
    in_expected_revision bigint,
    in_recovery_token_sha256 text,
    in_manifest_sha256 text,
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
AS $fn$
    UPDATE platform.tenancy_installation_deliveries AS delivery
    SET delivery_state = 'ready_to_display',
        delivery_revision = delivery.delivery_revision + 1,
        current_attempt_id = in_next_attempt_id,
        current_credential_id = in_next_credential_id,
        current_credential_version = in_next_credential_version,
        claim_adapter_id = NULL,
        claim_lease_until = NULL,
        recovery_cause = in_cause,
        delivery_manifest_sha256 = in_manifest_sha256,
        database_now = pg_catalog.clock_timestamp()
    WHERE delivery.operation_id = in_operation_id
      AND delivery.current_attempt_id = in_current_attempt_id
      AND delivery.current_credential_id = in_current_credential_id
      AND delivery.delivery_revision = in_expected_revision
    RETURNING delivery.delivery_state, delivery.delivery_revision, delivery.database_now;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_installation_delivery_record_v1(uuid, uuid, uuid, bigint, uuid, uuid, bigint, bigint, text, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_installation_delivery_record_v1(uuid, uuid, uuid, bigint, uuid, uuid, bigint, bigint, text, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_installation_delivery_record_v1(uuid, uuid, uuid, bigint, uuid, uuid, bigint, bigint, text, text, text);
