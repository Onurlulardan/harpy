-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_authority_ack_v1(
    p_operation_id uuid,
    p_registry_binding_id uuid,
    p_scope_kind text,
    p_scope_id uuid,
    p_generation bigint,
    p_previous_head_sha256 text,
    p_state_sha256 text,
    p_proposal_sha256 text,
    p_ack_sha256 text,
    p_registry_sequence bigint,
    p_verified_at timestamptz,
    p_expected_revision bigint
)
RETURNS TABLE (operation_id uuid, state text, revision bigint, updated_at timestamptz)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    operation platform.tenancy_onboarding_operations%ROWTYPE;
    existing platform.tenancy_onboarding_authority_acks%ROWTYPE;
    now_at timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR p_verified_at > now_at THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'authority acknowledgement denied';
    END IF;
    SELECT * INTO operation
    FROM platform.tenancy_onboarding_operations AS candidate
    WHERE candidate.operation_id = p_operation_id
    FOR UPDATE;
    IF NOT FOUND OR operation.deployment_mode <> 'saas' THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'SaaS onboarding operation required';
    END IF;
    SELECT * INTO existing
    FROM platform.tenancy_onboarding_authority_acks AS ack
    WHERE ack.operation_id = p_operation_id;
    IF FOUND THEN
        IF existing.registry_binding_id IS DISTINCT FROM p_registry_binding_id
           OR existing.scope_kind IS DISTINCT FROM p_scope_kind
           OR existing.scope_id IS DISTINCT FROM p_scope_id
           OR existing.generation IS DISTINCT FROM p_generation
           OR existing.previous_head_sha256 IS DISTINCT FROM p_previous_head_sha256
           OR existing.state_sha256 IS DISTINCT FROM p_state_sha256
           OR existing.proposal_sha256 IS DISTINCT FROM p_proposal_sha256
           OR existing.ack_sha256 IS DISTINCT FROM p_ack_sha256
           OR existing.registry_sequence IS DISTINCT FROM p_registry_sequence THEN
            RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'authority acknowledgement CAS conflict';
        END IF;
        RETURN QUERY SELECT operation.operation_id, operation.state,
            operation.revision, operation.updated_at;
        RETURN;
    END IF;
    IF operation.revision <> p_expected_revision OR operation.state <> 'authority_pending'
       OR operation.registry_binding_id IS DISTINCT FROM p_registry_binding_id
       OR operation.authority_scope_kind IS DISTINCT FROM p_scope_kind
       OR operation.authority_scope_id IS DISTINCT FROM p_scope_id
       OR operation.authority_generation IS DISTINCT FROM p_generation
       OR operation.authority_previous_head_sha256 IS DISTINCT FROM p_previous_head_sha256
       OR operation.authority_state_sha256 IS DISTINCT FROM p_state_sha256
       OR operation.authority_proposal_sha256 IS DISTINCT FROM p_proposal_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'authority plan CAS conflict';
    END IF;
    PERFORM 1 FROM platform.tenancy_onboarding_receipts AS receipt
    WHERE receipt.operation_id = p_operation_id
      AND receipt.receipt_kind = 'migration_completed';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'migration receipt required';
    END IF;
    INSERT INTO platform.tenancy_onboarding_authority_acks (
        operation_id, registry_binding_id, scope_kind, scope_id, generation,
        previous_head_sha256, state_sha256, proposal_sha256, ack_sha256,
        registry_sequence, verified_at
    ) VALUES (p_operation_id, p_registry_binding_id, p_scope_kind, p_scope_id,
        p_generation, p_previous_head_sha256, p_state_sha256, p_proposal_sha256,
        p_ack_sha256, p_registry_sequence, p_verified_at);
    UPDATE platform.tenancy_onboarding_operations
    SET state = 'authority_acknowledged', revision = operation.revision + 1,
        updated_at = now_at
    WHERE tenancy_onboarding_operations.operation_id = p_operation_id;
    RETURN QUERY SELECT p_operation_id, 'authority_acknowledged'::text,
        operation.revision + 1, now_at;
END;
$function$;
-- +goose StatementEnd
