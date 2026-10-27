-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_owner_transfer_compensate_apply_v1(
    p_tenant_id uuid, p_operation_id uuid, p_request_sha256 bytea,
    p_actor_id uuid, p_expected_platform_clock bigint,
    p_registry_binding_id uuid, p_scope_id uuid, p_registry_sequence bigint,
    p_generation bigint, p_proposal_sha256 bytea, p_previous_head_sha256 bytea,
    p_state_sha256 bytea, p_ack_sha256 bytea, p_ack_issued_at timestamptz,
    p_ack_registry_kid text
)
RETURNS TABLE (
    operation_id uuid, original_operation_id uuid, generation bigint,
    state_sha256 bytea, ack_sha256 bytea, replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_environment_id uuid; v_owner_subject_id uuid; v_clock bigint;
    v_original_operation_id uuid; v_original_generation bigint;
    v_original_previous_head bytea; v_compensates_ack bytea; v_status text;
    v_proposal_expires_at timestamptz; v_head_generation bigint; v_head_sha bytea;
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF p_operation_id IS NULL OR pg_catalog.octet_length(p_request_sha256) <> 32 OR
       p_actor_id IS NULL OR p_registry_binding_id IS NULL OR p_scope_id IS NULL OR
       p_registry_sequence < 1 OR p_generation < 3 OR
       pg_catalog.octet_length(p_proposal_sha256) <> 32 OR
       pg_catalog.octet_length(p_previous_head_sha256) <> 32 OR
       pg_catalog.octet_length(p_state_sha256) <> 32 OR
       pg_catalog.octet_length(p_ack_sha256) <> 32 OR p_ack_issued_at > v_now OR
       p_ack_registry_kid !~ '^[a-z0-9][a-z0-9._-]{0,127}$' THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'Owner compensation acknowledgement is invalid';
    END IF;
    SELECT context.environment_id, context.owner_subject_id, context.platform_clock
    INTO v_environment_id, v_owner_subject_id, v_clock
    FROM platform.iam_platform_owner_context_lock_v1(
        p_tenant_id, p_actor_id, p_expected_platform_clock
    ) AS context;
    SELECT compensation.original_operation_id, compensation.status,
           compensation.compensates_ack_sha256, compensation.proposal_expires_at
    INTO v_original_operation_id, v_status, v_compensates_ack, v_proposal_expires_at
    FROM platform.iam_platform_owner_transfer_compensations AS compensation
    WHERE compensation.operation_id = p_operation_id AND compensation.environment_id = v_environment_id
      AND compensation.request_sha256 = p_request_sha256 AND compensation.actor_user_id = p_actor_id
      AND compensation.registry_binding_id = p_registry_binding_id
      AND compensation.scope_kind = 'environment' AND compensation.scope_id = p_scope_id
      AND compensation.generation = p_generation
      AND compensation.proposal_sha256 = p_proposal_sha256
      AND compensation.previous_head_sha256 = p_previous_head_sha256
      AND compensation.state_sha256 = p_state_sha256 FOR UPDATE;
    IF v_status = 'completed' THEN
        IF NOT EXISTS (SELECT 1 FROM platform.iam_platform_owner_transfer_compensations AS compensation
            WHERE compensation.operation_id = p_operation_id AND compensation.ack_sha256 = p_ack_sha256
              AND compensation.registry_sequence = p_registry_sequence
              AND compensation.ack_registry_kid = p_ack_registry_kid) THEN
            RAISE EXCEPTION USING ERRCODE = 'P1401', MESSAGE = 'Owner compensation ack conflicts';
        END IF;
        RETURN QUERY SELECT p_operation_id,v_original_operation_id,p_generation,
            p_state_sha256,p_ack_sha256,true;
        RETURN;
    END IF;
    SELECT operation.generation, operation.previous_head_sha256
    INTO v_original_generation, v_original_previous_head
    FROM platform.iam_platform_owner_transfer_operations AS operation
    WHERE operation.operation_id = v_original_operation_id AND operation.status = 'ack_verified'
      AND operation.ack_sha256 = v_compensates_ack FOR UPDATE;
    SELECT head.generation, head.state_sha256 INTO v_head_generation, v_head_sha
    FROM platform.iam_protected_authority_heads AS head
    WHERE head.environment_id = v_environment_id FOR UPDATE;
    IF v_environment_id IS NULL OR v_status <> 'prepared' OR p_scope_id <> v_environment_id OR
       p_ack_issued_at > v_proposal_expires_at OR v_original_generation IS NULL OR
       p_generation <> v_original_generation + 1 OR
       v_head_generation <> v_original_generation - 1 OR
       v_head_sha <> v_original_previous_head THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'Owner compensation ABA guard failed';
    END IF;
    UPDATE platform.iam_protected_authority_heads AS head
    SET generation = p_generation, state_sha256 = p_state_sha256, ack_sha256 = p_ack_sha256,
        revision = head.revision + 1, updated_at = v_now
    WHERE head.environment_id = v_environment_id AND head.generation = v_original_generation - 1
      AND head.state_sha256 = v_original_previous_head;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'Owner compensation head changed';
    END IF;
    UPDATE platform.iam_platform_owner_transfer_compensations AS compensation
    SET status = 'completed', registry_sequence = p_registry_sequence,
        ack_sha256 = p_ack_sha256, ack_registry_kid = p_ack_registry_kid,
        ack_issued_at = p_ack_issued_at, revision = compensation.revision + 1,
        updated_at = v_now, completed_at = v_now
    WHERE compensation.operation_id = p_operation_id AND compensation.status = 'prepared';
    UPDATE platform.iam_platform_owner_transfer_operations AS operation
    SET status = 'compensated', revision = operation.revision + 1,
        updated_at = v_now, completed_at = v_now
    WHERE operation.operation_id = v_original_operation_id AND operation.status = 'ack_verified';
    RETURN QUERY SELECT p_operation_id,v_original_operation_id,p_generation,
        p_state_sha256,p_ack_sha256,false;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.iam_owner_transfer_compensate_apply_v1(
    uuid,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint,bytea,bytea,bytea,
    bytea,timestamptz,text
) FROM PUBLIC;
