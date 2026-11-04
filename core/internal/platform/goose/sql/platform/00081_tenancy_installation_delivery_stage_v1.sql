-- +goose Up
CREATE FUNCTION platform.tenancy_installation_delivery_stage_v1(
    in_operation_id uuid,
    in_attempt_id uuid,
    in_credential_id uuid,
    in_credential_version bigint,
    in_installation_manifest_sha256 text,
    in_delivery_manifest_sha256 text
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
AS $$
    WITH identity AS (
        SELECT operation.tenant_id, operation.first_owner_user_id
        FROM platform.tenancy_onboarding_operations AS operation
        WHERE operation.operation_id = in_operation_id
    ), inserted AS (
        INSERT INTO platform.tenancy_installation_deliveries (
            operation_id, tenant_id, first_owner_user_id, current_attempt_id,
            current_credential_id, current_credential_version, delivery_state,
            delivery_revision, database_now, installation_manifest_sha256,
            delivery_manifest_sha256
        )
        SELECT in_operation_id, identity.tenant_id, identity.first_owner_user_id,
               in_attempt_id, in_credential_id, in_credential_version,
               'ready_to_display', 1, pg_catalog.clock_timestamp(),
               in_installation_manifest_sha256, in_delivery_manifest_sha256
        FROM identity
        ON CONFLICT (operation_id) DO NOTHING
        RETURNING operation_id
    )
    SELECT delivery.operation_id, delivery.tenant_id, delivery.first_owner_user_id,
           delivery.current_attempt_id, delivery.current_credential_id,
           delivery.current_credential_version, delivery.delivery_state,
           delivery.delivery_revision, delivery.database_now, delivery.claim_adapter_id,
           delivery.claim_lease_until, delivery.acknowledged_at, delivery.secret_deleted_at,
           delivery.completed_at
    FROM platform.tenancy_installation_deliveries AS delivery
    WHERE delivery.operation_id = in_operation_id
      AND EXISTS (SELECT 1 FROM identity);
$$;

REVOKE ALL ON FUNCTION platform.tenancy_installation_delivery_stage_v1(uuid, uuid, uuid, bigint, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_installation_delivery_stage_v1(uuid, uuid, uuid, bigint, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_installation_delivery_stage_v1(uuid, uuid, uuid, bigint, text, text);
