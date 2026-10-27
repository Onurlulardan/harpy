-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_reconcile_v1(
    p_operation_id uuid,
    p_receipt_id uuid,
    p_receipt_sha256 text,
    p_evidence_sha256 text,
    p_expected_revision bigint
)
RETURNS TABLE (operation_id uuid, state text, revision bigint, updated_at timestamptz)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    operation platform.tenancy_onboarding_operations%ROWTYPE;
    existing platform.tenancy_onboarding_receipts%ROWTYPE;
    current_lease platform.tenancy_onboarding_leases%ROWTYPE;
    latest_attempt_id uuid;
    next_state text;
    now_at timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'onboarding reconcile denied';
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
        IF existing.operation_id IS DISTINCT FROM p_operation_id
           OR existing.receipt_kind <> 'ambiguous_reconciled'
           OR existing.receipt_sha256 IS DISTINCT FROM p_receipt_sha256
           OR existing.evidence_sha256 IS DISTINCT FROM p_evidence_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = '23505', MESSAGE = 'reconcile receipt replay conflict';
        END IF;
        RETURN QUERY SELECT operation.operation_id, operation.state,
            operation.revision, operation.updated_at;
        RETURN;
    END IF;
    IF operation.revision <> p_expected_revision OR operation.state = 'fenced' THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'onboarding reconcile revision conflict';
    END IF;
    SELECT * INTO current_lease
    FROM platform.tenancy_onboarding_leases AS lease
    WHERE lease.operation_id = p_operation_id
    FOR UPDATE;
    IF FOUND AND current_lease.lease_until <= now_at THEN
        UPDATE platform.tenancy_onboarding_attempts
        SET outcome = 'ambiguous', finished_at = now_at
        WHERE attempt_id = current_lease.attempt_id AND outcome = 'active';
        DELETE FROM platform.tenancy_onboarding_leases AS expired_lease
        WHERE expired_lease.operation_id = p_operation_id;
    ELSIF FOUND AND current_lease.lease_until > now_at THEN
        RAISE EXCEPTION USING ERRCODE = '55P03', MESSAGE = 'active lease cannot be reconciled';
    END IF;
    SELECT attempt.attempt_id INTO latest_attempt_id
    FROM platform.tenancy_onboarding_attempts AS attempt
    WHERE attempt.operation_id = p_operation_id
    ORDER BY attempt.attempt_number DESC LIMIT 1;
    IF latest_attempt_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'onboarding attempt required';
    END IF;
    IF EXISTS (SELECT 1 FROM platform.tenancy_onboarding_receipts AS receipt
        WHERE receipt.operation_id = p_operation_id
          AND receipt.receipt_kind = 'activation_completed') THEN
        next_state := 'succeeded';
    ELSIF EXISTS (SELECT 1 FROM platform.tenancy_onboarding_authority_acks AS ack
        WHERE ack.operation_id = p_operation_id) THEN
        next_state := 'authority_acknowledged';
    ELSIF EXISTS (SELECT 1 FROM platform.tenancy_onboarding_receipts AS receipt
        WHERE receipt.operation_id = p_operation_id
          AND receipt.receipt_kind = 'migration_completed') THEN
        next_state := CASE operation.deployment_mode
            WHEN 'saas' THEN 'authority_pending' ELSE 'migration_ready' END;
    ELSE
        next_state := 'accepted';
    END IF;
    INSERT INTO platform.tenancy_onboarding_receipts (
        receipt_id, operation_id, attempt_id, receipt_kind,
        receipt_sha256, evidence_sha256, occurred_at
    ) VALUES (p_receipt_id, p_operation_id, latest_attempt_id,
        'ambiguous_reconciled', p_receipt_sha256, p_evidence_sha256, now_at);
    UPDATE platform.tenancy_onboarding_operations
    SET state = next_state, revision = operation.revision + 1, updated_at = now_at
    WHERE tenancy_onboarding_operations.operation_id = p_operation_id;
    RETURN QUERY SELECT p_operation_id, next_state, operation.revision + 1, now_at;
END;
$function$;
-- +goose StatementEnd
