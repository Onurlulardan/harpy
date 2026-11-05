-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION platform.tenancy_installation_delivery_read_v1(
    in_operation_id uuid
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
    SELECT delivery.operation_id, delivery.tenant_id, delivery.first_owner_user_id,
           delivery.current_attempt_id, delivery.current_credential_id,
           delivery.current_credential_version, delivery.delivery_state,
           delivery.delivery_revision, delivery.database_now, delivery.claim_adapter_id,
           delivery.claim_lease_until, delivery.acknowledged_at, delivery.secret_deleted_at,
           delivery.completed_at
    FROM platform.tenancy_installation_deliveries AS delivery
    WHERE delivery.operation_id = in_operation_id
      AND delivery.operation_id IS NOT NULL
      AND delivery.tenant_id IS NOT NULL
      AND delivery.operation_id <> '00000000-0000-0000-0000-000000000000'::uuid
      AND delivery.tenant_id <> '00000000-0000-0000-0000-000000000000'::uuid;
$fn$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION platform.tenancy_installation_delivery_read_v1(
    in_operation_id uuid
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
    SELECT delivery.operation_id, delivery.tenant_id, delivery.first_owner_user_id,
           delivery.current_attempt_id, delivery.current_credential_id,
           delivery.current_credential_version, delivery.delivery_state,
           delivery.delivery_revision, delivery.database_now, delivery.claim_adapter_id,
           delivery.claim_lease_until, delivery.acknowledged_at, delivery.secret_deleted_at,
           delivery.completed_at
    FROM platform.tenancy_installation_deliveries AS delivery
    WHERE delivery.operation_id = in_operation_id;
$fn$;
-- +goose StatementEnd
