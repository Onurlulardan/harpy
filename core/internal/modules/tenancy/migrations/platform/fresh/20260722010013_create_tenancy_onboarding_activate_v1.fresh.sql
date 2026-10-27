-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_activate_v1(
    p_operation_id uuid,
    p_lease_token_sha256 text,
    p_expected_revision bigint
)
RETURNS TABLE (
    operation_id uuid,
    tenant_id uuid,
    operation_kind text,
    step_plan_sha256 text,
    migration_receipt_id uuid,
    migration_receipt_sha256 text,
    authority_ack_sha256 text,
    revision bigint,
    database_now timestamptz,
    lease_until timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    operation platform.tenancy_onboarding_operations%ROWTYPE;
    lease platform.tenancy_onboarding_leases%ROWTYPE;
    migration_receipt platform.tenancy_onboarding_receipts%ROWTYPE;
    ack_sha256 text;
    now_at timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'tenant activation claim denied';
    END IF;
    SELECT * INTO operation
    FROM platform.tenancy_onboarding_operations AS candidate
    WHERE candidate.operation_id = p_operation_id
    FOR UPDATE;
    IF NOT FOUND OR operation.revision <> p_expected_revision
       OR NOT (operation.deployment_mode = 'onpremise' AND operation.state = 'migration_ready'
           OR operation.deployment_mode = 'saas' AND operation.state = 'authority_acknowledged') THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'tenant activation state conflict';
    END IF;
    SELECT * INTO lease
    FROM platform.tenancy_onboarding_leases AS current_lease
    WHERE current_lease.operation_id = p_operation_id
      AND current_lease.lease_token_sha256 = p_lease_token_sha256
      AND current_lease.lease_until > now_at
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'active activation lease required';
    END IF;
    SELECT * INTO migration_receipt
    FROM platform.tenancy_onboarding_receipts AS receipt
    WHERE receipt.operation_id = p_operation_id
      AND receipt.receipt_kind = 'migration_completed';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'migration completion receipt required';
    END IF;
    SELECT ack.ack_sha256 INTO ack_sha256
    FROM platform.tenancy_onboarding_authority_acks AS ack
    WHERE ack.operation_id = p_operation_id;
    IF operation.deployment_mode = 'saas' AND ack_sha256 IS NULL
       OR operation.deployment_mode = 'onpremise' AND ack_sha256 IS NOT NULL THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'authority acknowledgement state mismatch';
    END IF;
    UPDATE platform.tenancy_onboarding_operations
    SET state = 'finalizing', revision = operation.revision + 1, updated_at = now_at
    WHERE tenancy_onboarding_operations.operation_id = p_operation_id;
    UPDATE platform.tenancy_onboarding_leases
    SET revision = operation.revision + 1, heartbeat_at = now_at
    WHERE tenancy_onboarding_leases.operation_id = p_operation_id;
    RETURN QUERY SELECT operation.operation_id, operation.tenant_id,
        operation.operation_kind, operation.step_plan_sha256,
        migration_receipt.receipt_id, migration_receipt.receipt_sha256,
        ack_sha256, operation.revision + 1, now_at, lease.lease_until;
END;
$function$;
-- +goose StatementEnd
