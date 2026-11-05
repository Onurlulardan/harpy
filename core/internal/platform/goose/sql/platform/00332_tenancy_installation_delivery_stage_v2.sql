-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION platform.tenancy_installation_delivery_stage_v1(
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
AS $fn$
    WITH identity AS (
        SELECT operation.tenant_id, operation.first_owner_user_id
        FROM platform.tenancy_onboarding_operations AS operation
        WHERE operation.operation_id = in_operation_id
          AND operation.tenant_id <> '00000000-0000-0000-0000-000000000000'::uuid
          AND operation.first_owner_user_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ), stored AS (
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
        ON CONFLICT (operation_id) DO UPDATE
        SET delivery_revision = platform.tenancy_installation_deliveries.delivery_revision
        RETURNING operation_id, tenant_id, first_owner_user_id, current_attempt_id,
                  current_credential_id, current_credential_version, delivery_state,
                  delivery_revision, database_now, claim_adapter_id, claim_lease_until,
                  acknowledged_at, secret_deleted_at, completed_at
    )
    SELECT stored.operation_id, stored.tenant_id, stored.first_owner_user_id,
           stored.current_attempt_id, stored.current_credential_id,
           stored.current_credential_version, stored.delivery_state,
           stored.delivery_revision, stored.database_now, stored.claim_adapter_id,
           stored.claim_lease_until, stored.acknowledged_at, stored.secret_deleted_at,
           stored.completed_at
    FROM stored
    WHERE stored.operation_id IS NOT NULL
      AND stored.operation_id <> '00000000-0000-0000-0000-000000000000'::uuid
      AND stored.tenant_id <> '00000000-0000-0000-0000-000000000000'::uuid;
$fn$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION platform.tenancy_installation_delivery_stage_v1(
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
AS $fn$
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
$fn$;
-- +goose StatementEnd
