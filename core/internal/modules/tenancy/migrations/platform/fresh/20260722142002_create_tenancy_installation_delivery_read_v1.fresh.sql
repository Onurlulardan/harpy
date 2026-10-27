-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_installation_delivery_read_v1(p_operation_id uuid)
RETURNS TABLE (
    operation_id uuid, tenant_id uuid, first_owner_user_id uuid,
    current_attempt_id uuid, current_credential_id uuid,
    current_credential_version bigint, delivery_state text,
    delivery_revision bigint, database_now timestamptz,
    claim_adapter_id text, claim_lease_until timestamptz,
    acknowledged_at timestamptz, secret_deleted_at timestamptz,
    completed_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501', MESSAGE='installation delivery read denied';
    END IF;
    RETURN QUERY SELECT delivery.operation_id,delivery.tenant_id,delivery.first_owner_user_id,
        delivery.current_attempt_id,delivery.current_credential_id,
        delivery.current_credential_version,delivery.state,delivery.revision,
        pg_catalog.statement_timestamp(),delivery.claim_adapter_id,
        delivery.claim_lease_until,delivery.acknowledged_at,
        delivery.secret_deleted_at,delivery.completed_at
    FROM platform.tenancy_installation_credential_deliveries AS delivery
    WHERE delivery.operation_id=p_operation_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_installation_delivery_read_v1(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_installation_delivery_read_v1(uuid)
    TO harpy_runtime;
