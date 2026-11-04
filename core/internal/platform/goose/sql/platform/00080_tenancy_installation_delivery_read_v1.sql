-- +goose Up
CREATE TABLE platform.tenancy_onboarding_operations (
    operation_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    first_owner_user_id uuid NOT NULL,
    kind text NOT NULL,
    step_plan_sha256 text NOT NULL,
    deployment_mode text NOT NULL,
    state text NOT NULL,
    revision bigint NOT NULL,
    attempt_number bigint NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE platform.tenancy_installation_deliveries (
    operation_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    first_owner_user_id uuid NOT NULL,
    current_attempt_id uuid NOT NULL,
    current_credential_id uuid NOT NULL,
    current_credential_version bigint NOT NULL,
    delivery_state text NOT NULL,
    delivery_revision bigint NOT NULL,
    database_now timestamptz NOT NULL,
    claim_adapter_id text,
    claim_lease_until timestamptz,
    acknowledged_at timestamptz,
    secret_deleted_at timestamptz,
    completed_at timestamptz,
    installation_manifest_sha256 text,
    delivery_manifest_sha256 text,
    recovery_cause text
);

REVOKE ALL ON TABLE platform.tenancy_onboarding_operations FROM PUBLIC;
REVOKE ALL ON TABLE platform.tenancy_installation_deliveries FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE ON TABLE platform.tenancy_onboarding_operations TO harpy_api, harpy_worker;
GRANT SELECT, INSERT, UPDATE ON TABLE platform.tenancy_installation_deliveries TO harpy_api, harpy_worker;

CREATE FUNCTION platform.tenancy_installation_delivery_read_v1(
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
AS $$
    SELECT delivery.operation_id, delivery.tenant_id, delivery.first_owner_user_id,
           delivery.current_attempt_id, delivery.current_credential_id,
           delivery.current_credential_version, delivery.delivery_state,
           delivery.delivery_revision, delivery.database_now, delivery.claim_adapter_id,
           delivery.claim_lease_until, delivery.acknowledged_at, delivery.secret_deleted_at,
           delivery.completed_at
    FROM platform.tenancy_installation_deliveries AS delivery
    WHERE delivery.operation_id = in_operation_id;
$$;

REVOKE ALL ON FUNCTION platform.tenancy_installation_delivery_read_v1(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_installation_delivery_read_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_installation_delivery_read_v1(uuid);
DROP TABLE platform.tenancy_installation_deliveries;
DROP TABLE platform.tenancy_onboarding_operations;
