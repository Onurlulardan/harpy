-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_owner_transfer_ack_v1(
    p_tenant_id uuid, p_operation_id uuid, p_request_sha256 bytea,
    p_actor_id uuid, p_expected_platform_clock bigint,
    p_registry_binding_id uuid, p_scope_id uuid, p_registry_sequence bigint,
    p_generation bigint, p_proposal_sha256 bytea, p_previous_head_sha256 bytea,
    p_state_sha256 bytea, p_ack_sha256 bytea, p_ack_issued_at timestamptz,
    p_ack_registry_kid text
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
    v_status text; v_operation_environment uuid; v_operation_actor uuid;
    v_operation_request bytea; v_operation_registry uuid; v_operation_scope uuid;
    v_operation_generation bigint; v_operation_proposal bytea;
    v_operation_previous bytea; v_operation_state bytea; v_operation_ack bytea;
    v_operation_sequence bigint; v_operation_ack_kid text; v_operation_expires_at timestamptz;
    v_head_generation bigint; v_head_sha bytea; v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF p_operation_id IS NULL OR pg_catalog.octet_length(p_request_sha256) <> 32 OR
       p_actor_id IS NULL OR p_registry_binding_id IS NULL OR p_scope_id IS NULL OR
       p_registry_sequence < 1 OR p_generation < 2 OR
       pg_catalog.octet_length(p_proposal_sha256) <> 32 OR
       pg_catalog.octet_length(p_previous_head_sha256) <> 32 OR
       pg_catalog.octet_length(p_state_sha256) <> 32 OR
       pg_catalog.octet_length(p_ack_sha256) <> 32 OR p_ack_issued_at > v_now OR
       p_ack_registry_kid !~ '^[a-z0-9][a-z0-9._-]{0,127}$' THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'Owner transfer acknowledgement is invalid';
    END IF;
    SELECT context.environment_id, context.owner_subject_id, context.platform_clock
    INTO v_environment_id, v_owner_subject_id, v_clock
    FROM platform.iam_platform_owner_context_lock_v1(
        p_tenant_id, p_actor_id, p_expected_platform_clock
    ) AS context;
    IF v_environment_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Current actor is not the Instance Owner';
    END IF;
    SELECT operation.status,operation.environment_id,operation.actor_user_id,
           operation.request_sha256,operation.registry_binding_id,operation.scope_id,
           operation.generation,operation.proposal_sha256,operation.previous_head_sha256,
           operation.state_sha256,operation.ack_sha256,operation.registry_sequence,
           operation.ack_registry_kid,operation.proposal_expires_at
    INTO v_status,v_operation_environment,v_operation_actor,v_operation_request,
         v_operation_registry,v_operation_scope,v_operation_generation,v_operation_proposal,
         v_operation_previous,v_operation_state,v_operation_ack,v_operation_sequence,
         v_operation_ack_kid,v_operation_expires_at
    FROM platform.iam_platform_owner_transfer_operations AS operation
    WHERE operation.operation_id = p_operation_id FOR UPDATE;
    IF v_status = 'ack_verified' THEN
        IF v_operation_request <> p_request_sha256 OR v_operation_ack <> p_ack_sha256 OR
           v_operation_sequence <> p_registry_sequence OR v_operation_ack_kid <> p_ack_registry_kid THEN
            RAISE EXCEPTION USING ERRCODE = 'P1401', MESSAGE = 'Owner transfer acknowledgement conflicts';
        END IF;
        RETURN QUERY SELECT p_operation_id,v_operation_registry,v_operation_scope,
            v_operation_generation,v_operation_previous,v_operation_state,
            v_operation_proposal,v_status,true;
        RETURN;
    END IF;
    SELECT head.generation, head.state_sha256 INTO v_head_generation, v_head_sha
    FROM platform.iam_protected_authority_heads AS head
    WHERE head.environment_id = v_environment_id FOR UPDATE;
    IF v_status IS NULL OR v_status <> 'prepared' OR
       v_operation_environment <> v_environment_id OR v_operation_actor <> p_actor_id OR
       v_operation_request <> p_request_sha256 OR v_operation_registry <> p_registry_binding_id OR
       v_operation_scope <> p_scope_id OR p_scope_id <> v_environment_id OR
       v_operation_generation <> p_generation OR v_operation_proposal <> p_proposal_sha256 OR
       v_operation_previous <> p_previous_head_sha256 OR v_operation_state <> p_state_sha256 OR
       p_ack_issued_at > v_operation_expires_at OR
       v_head_generation <> p_generation - 1 OR v_head_sha <> p_previous_head_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'Owner transfer acknowledgement CAS failed';
    END IF;
    UPDATE platform.iam_platform_owner_transfer_operations AS operation
    SET status = 'ack_verified', registry_sequence = p_registry_sequence,
        ack_sha256 = p_ack_sha256, ack_registry_kid = p_ack_registry_kid,
        ack_issued_at = p_ack_issued_at, revision = operation.revision + 1, updated_at = v_now
    WHERE operation.operation_id = p_operation_id AND operation.status = 'prepared';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'Owner transfer acknowledgement changed';
    END IF;
    RETURN QUERY SELECT p_operation_id,v_operation_registry,v_operation_scope,
        v_operation_generation,v_operation_previous,v_operation_state,
        v_operation_proposal,'ack_verified'::text,false;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.iam_owner_transfer_ack_v1(
    uuid,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint,bytea,bytea,bytea,
    bytea,timestamptz,text
) FROM PUBLIC;
