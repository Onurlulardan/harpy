-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_owner_transfer_compensate_prepare_v1(
    p_tenant_id uuid, p_operation_id uuid, p_request_sha256 bytea,
    p_actor_id uuid, p_expected_platform_clock bigint, p_original_operation_id uuid,
    p_original_ack_sha256 bytea, p_expected_original_generation bigint,
    p_proposal_sha256 bytea, p_proposal_signer_kid text,
    p_proposal_signature_sha256 bytea, p_proposal_expires_at timestamptz
)
RETURNS TABLE (
    operation_id uuid, registry_binding_id uuid, scope_id uuid, generation bigint,
    previous_head_sha256 bytea, state_sha256 bytea, proposal_sha256 bytea,
    status text, replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_environment_id uuid; v_owner_subject_id uuid; v_clock bigint;
    v_registry_binding_id uuid; v_original_generation bigint;
    v_original_previous_head bytea; v_original_state bytea; v_original_status text;
    v_head_generation bigint; v_head_sha bytea; v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF p_operation_id IS NULL OR p_operation_id = p_original_operation_id OR
       pg_catalog.octet_length(p_request_sha256) <> 32 OR p_actor_id IS NULL OR
       pg_catalog.octet_length(p_original_ack_sha256) <> 32 OR
       p_expected_original_generation < 2 OR pg_catalog.octet_length(p_proposal_sha256) <> 32 OR
       p_proposal_signer_kid !~ '^[a-z0-9][a-z0-9._-]{0,127}$' OR
       pg_catalog.octet_length(p_proposal_signature_sha256) <> 32 OR
       p_proposal_expires_at < v_now OR p_proposal_expires_at > v_now + INTERVAL '7 days' THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'Owner transfer compensation is invalid';
    END IF;
    IF EXISTS (SELECT 1 FROM platform.iam_platform_owner_transfer_compensations AS compensation
               WHERE compensation.operation_id = p_operation_id) THEN
        IF NOT EXISTS (SELECT 1 FROM platform.iam_platform_owner_transfer_compensations AS compensation
            WHERE compensation.operation_id = p_operation_id
              AND compensation.original_operation_id = p_original_operation_id
              AND compensation.request_sha256 = p_request_sha256
              AND compensation.proposal_sha256 = p_proposal_sha256
              AND compensation.status = 'prepared') THEN
            RAISE EXCEPTION USING ERRCODE = 'P1401', MESSAGE = 'Owner compensation idempotency conflict';
        END IF;
        RETURN QUERY SELECT compensation.operation_id, compensation.registry_binding_id,
            compensation.scope_id, compensation.generation, compensation.previous_head_sha256,
            compensation.state_sha256, compensation.proposal_sha256, compensation.status, true
        FROM platform.iam_platform_owner_transfer_compensations AS compensation
        WHERE compensation.operation_id = p_operation_id;
        RETURN;
    END IF;
    SELECT context.environment_id, context.owner_subject_id, context.platform_clock
    INTO v_environment_id, v_owner_subject_id, v_clock
    FROM platform.iam_platform_owner_context_lock_v1(
        p_tenant_id, p_actor_id, p_expected_platform_clock
    ) AS context;
    SELECT operation.registry_binding_id, operation.generation,
           operation.previous_head_sha256, operation.state_sha256, operation.status
    INTO v_registry_binding_id, v_original_generation,
         v_original_previous_head, v_original_state, v_original_status
    FROM platform.iam_platform_owner_transfer_operations AS operation
    WHERE operation.operation_id = p_original_operation_id
      AND operation.environment_id = v_environment_id
      AND operation.actor_user_id = p_actor_id
      AND operation.ack_sha256 = p_original_ack_sha256 FOR UPDATE;
    SELECT head.generation, head.state_sha256 INTO v_head_generation, v_head_sha
    FROM platform.iam_protected_authority_heads AS head
    WHERE head.environment_id = v_environment_id FOR UPDATE;
    IF v_environment_id IS NULL OR v_original_status <> 'ack_verified' OR
       v_original_generation <> p_expected_original_generation OR
       v_head_generation <> v_original_generation - 1 OR
       v_head_sha <> v_original_previous_head THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'Owner compensation CAS failed';
    END IF;
    INSERT INTO platform.iam_platform_owner_transfer_compensations (
        operation_id,original_operation_id,environment_id,request_sha256,actor_user_id,
        expected_platform_clock,registry_binding_id,scope_kind,scope_id,generation,
        previous_head_sha256,state_sha256,proposal_sha256,proposal_signer_kid,
        proposal_signature_sha256,proposal_expires_at,compensates_ack_sha256,
        status,revision,created_at,updated_at
    ) VALUES (p_operation_id,p_original_operation_id,v_environment_id,p_request_sha256,p_actor_id,
        p_expected_platform_clock,v_registry_binding_id,'environment',v_environment_id,
        v_original_generation + 1,v_original_state,v_original_previous_head,p_proposal_sha256,
        p_proposal_signer_kid,p_proposal_signature_sha256,p_proposal_expires_at,
        p_original_ack_sha256,'prepared',1,v_now,v_now);
    RETURN QUERY SELECT p_operation_id,v_registry_binding_id,v_environment_id,
        v_original_generation + 1,v_original_state,v_original_previous_head,
        p_proposal_sha256,'prepared'::text,false;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.iam_owner_transfer_compensate_prepare_v1(
    uuid,uuid,bytea,uuid,bigint,uuid,bytea,bigint,bytea,text,bytea,timestamptz
) FROM PUBLIC;
