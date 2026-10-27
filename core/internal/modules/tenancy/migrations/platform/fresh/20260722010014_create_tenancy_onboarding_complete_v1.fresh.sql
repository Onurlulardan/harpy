-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_complete_v1(
    p_operation_id uuid,
    p_receipt_id uuid,
    p_receipt_sha256 text,
    p_evidence_sha256 text,
    p_expected_revision bigint
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    operation platform.tenancy_onboarding_operations%ROWTYPE;
    lease platform.tenancy_onboarding_leases%ROWTYPE;
    existing platform.tenancy_onboarding_receipts%ROWTYPE;
    now_at timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'tenant activation completion denied';
    END IF;
    SELECT * INTO operation
    FROM platform.tenancy_onboarding_operations AS candidate
    WHERE candidate.operation_id = p_operation_id
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002', MESSAGE = 'onboarding operation missing';
    END IF;
    SELECT * INTO existing
    FROM platform.tenancy_onboarding_receipts AS receipt
    WHERE receipt.receipt_id = p_receipt_id;
    IF FOUND THEN
        IF operation.state <> 'succeeded'
           OR existing.operation_id IS DISTINCT FROM p_operation_id
           OR existing.receipt_kind <> 'activation_completed'
           OR existing.receipt_sha256 IS DISTINCT FROM p_receipt_sha256
           OR existing.evidence_sha256 IS DISTINCT FROM p_evidence_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = '23505', MESSAGE = 'activation completion replay conflict';
        END IF;
        RETURN;
    END IF;
    IF operation.state <> 'finalizing' OR operation.revision <> p_expected_revision THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'activation completion revision conflict';
    END IF;
    SELECT * INTO lease
    FROM platform.tenancy_onboarding_leases AS current_lease
    WHERE current_lease.operation_id = p_operation_id
      AND current_lease.lease_until > now_at
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'active activation lease required';
    END IF;
    INSERT INTO platform.tenancy_onboarding_receipts (
        receipt_id, operation_id, attempt_id, receipt_kind,
        receipt_sha256, evidence_sha256, occurred_at
    ) VALUES (p_receipt_id, p_operation_id, lease.attempt_id,
        'activation_completed', p_receipt_sha256, p_evidence_sha256, now_at);
    UPDATE platform.tenancy_onboarding_attempts
    SET outcome = 'completed', finished_at = now_at
    WHERE attempt_id = lease.attempt_id AND outcome = 'active';
    DELETE FROM platform.tenancy_onboarding_leases WHERE operation_id = p_operation_id;
    UPDATE platform.tenancy_onboarding_operations
    SET state = 'succeeded', revision = operation.revision + 1, updated_at = now_at
    WHERE tenancy_onboarding_operations.operation_id = p_operation_id;
END;
$function$;
-- +goose StatementEnd
