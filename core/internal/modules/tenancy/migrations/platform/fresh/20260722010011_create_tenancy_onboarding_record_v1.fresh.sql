-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_record_v1(
    p_receipt_id uuid,
    p_operation_id uuid,
    p_attempt_id uuid,
    p_receipt_kind text,
    p_receipt_sha256 text,
    p_evidence_sha256 text,
    p_occurred_at timestamptz,
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
    next_state text;
    now_at timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR p_receipt_kind NOT IN ('migration_claim_ready', 'migration_completed', 'failure')
       OR p_occurred_at > now_at THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'onboarding receipt denied';
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
           OR existing.attempt_id IS DISTINCT FROM p_attempt_id
           OR existing.receipt_kind IS DISTINCT FROM p_receipt_kind
           OR existing.receipt_sha256 IS DISTINCT FROM p_receipt_sha256
           OR existing.evidence_sha256 IS DISTINCT FROM p_evidence_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = '23505', MESSAGE = 'onboarding receipt replay conflict';
        END IF;
        RETURN QUERY SELECT operation.operation_id, operation.state,
            operation.revision, operation.updated_at;
        RETURN;
    END IF;
    IF operation.revision <> p_expected_revision OR operation.state <> 'migrating' THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'onboarding receipt revision conflict';
    END IF;
    PERFORM 1 FROM platform.tenancy_onboarding_leases AS lease
    WHERE lease.operation_id = p_operation_id AND lease.attempt_id = p_attempt_id
      AND lease.lease_until > now_at
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'active onboarding lease required';
    END IF;
    INSERT INTO platform.tenancy_onboarding_receipts (
        receipt_id, operation_id, attempt_id, receipt_kind,
        receipt_sha256, evidence_sha256, occurred_at
    ) VALUES (p_receipt_id, p_operation_id, p_attempt_id, p_receipt_kind,
        p_receipt_sha256, p_evidence_sha256, p_occurred_at);
    next_state := CASE p_receipt_kind
        WHEN 'migration_completed' THEN CASE operation.deployment_mode
            WHEN 'saas' THEN 'authority_pending' ELSE 'migration_ready' END
        WHEN 'failure' THEN 'failed'
        ELSE 'migrating'
    END;
    UPDATE platform.tenancy_onboarding_operations
    SET state = next_state, revision = operation.revision + 1, updated_at = now_at
    WHERE tenancy_onboarding_operations.operation_id = p_operation_id;
    IF p_receipt_kind = 'failure' THEN
        UPDATE platform.tenancy_onboarding_attempts
        SET outcome = 'failed', finished_at = now_at
        WHERE attempt_id = p_attempt_id AND outcome = 'active';
        DELETE FROM platform.tenancy_onboarding_leases WHERE operation_id = p_operation_id;
    END IF;
    RETURN QUERY SELECT p_operation_id, next_state, operation.revision + 1, now_at;
END;
$function$;
-- +goose StatementEnd
