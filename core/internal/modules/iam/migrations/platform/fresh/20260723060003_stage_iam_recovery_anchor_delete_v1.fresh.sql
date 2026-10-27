-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_recovery_anchor_stage_delete_v1(
    p_operation_id uuid,p_environment_id uuid,p_tenant_id uuid,p_actor_user_id uuid,
    p_anchor_id uuid,p_idempotency_key text,p_request_sha256 bytea,p_occurred_at timestamptz
)
RETURNS TABLE(anchor_id uuid,anchor_set_sha256 bytea,replayed boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_existing platform.iam_recovery_anchor_operations%ROWTYPE; v_set bytea;
BEGIN
    IF p_operation_id IS NULL OR p_environment_id IS NULL OR p_tenant_id IS NULL OR
       p_actor_user_id IS NULL OR p_anchor_id IS NULL OR
       p_idempotency_key !~ '^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$' OR
       pg_catalog.octet_length(p_request_sha256)<>32 OR p_occurred_at IS NULL OR
       NOT platform.iam_platform_current_owner_v1(p_tenant_id,p_actor_user_id) THEN
        RAISE EXCEPTION USING ERRCODE='22023',MESSAGE='recovery anchor deletion is invalid';
    END IF;
    SELECT * INTO v_existing FROM platform.iam_recovery_anchor_operations AS operation
     WHERE operation.operation_id=p_operation_id FOR UPDATE;
    IF FOUND THEN
        IF v_existing.environment_id<>p_environment_id OR v_existing.tenant_id<>p_tenant_id OR
           v_existing.actor_user_id<>p_actor_user_id OR v_existing.kind<>'anchor_delete' OR
           v_existing.anchor_id<>p_anchor_id OR
           v_existing.options_idempotency_key<>p_idempotency_key OR
           v_existing.options_request_sha256<>p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE='23505',MESSAGE='recovery anchor deletion conflicts';
        END IF;
        RETURN QUERY SELECT p_anchor_id,v_existing.anchor_set_sha256,true;
        RETURN;
    END IF;
    IF NOT EXISTS(SELECT 1 FROM platform.iam_recovery_anchors AS anchor
        WHERE anchor.anchor_id=p_anchor_id AND anchor.environment_id=p_environment_id
          AND anchor.tenant_id=p_tenant_id AND anchor.state='active' FOR UPDATE) OR
       (SELECT COUNT(*) FROM platform.iam_recovery_anchors AS anchor
         WHERE anchor.environment_id=p_environment_id AND anchor.state='active')<=1 THEN
        RAISE EXCEPTION USING ERRCODE='40001',MESSAGE='last recovery anchor cannot be deleted';
    END IF;
    v_set:=platform.iam_recovery_anchor_set_sha256_v1(p_environment_id,p_anchor_id);
    INSERT INTO platform.iam_recovery_anchor_operations(
        operation_id,environment_id,tenant_id,actor_user_id,kind,anchor_id,
        options_idempotency_key,options_request_sha256,anchor_set_sha256,state,created_at,updated_at
    ) VALUES(p_operation_id,p_environment_id,p_tenant_id,p_actor_user_id,'anchor_delete',p_anchor_id,
        p_idempotency_key,p_request_sha256,v_set,'authority_pending',p_occurred_at,p_occurred_at);
    IF NOT platform.platformops_record_anchor_operation_v1(
        p_operation_id,p_environment_id,p_actor_user_id,p_idempotency_key,
        p_request_sha256,'anchor_delete',p_occurred_at) THEN
        RAISE EXCEPTION USING ERRCODE='40001',MESSAGE='recovery operation projection conflicts';
    END IF;
    RETURN QUERY SELECT p_anchor_id,v_set,false;
END;
$function$;

CREATE FUNCTION platform.iam_recovery_anchor_record_proposal_v1(
    p_operation_id uuid,p_tenant_id uuid,p_actor_user_id uuid,
    p_registry_binding_id uuid,p_authority_generation bigint,
    p_protected_state_sha256 bytea,p_proposal_sha256 bytea,p_occurred_at timestamptz
)
RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_operation platform.iam_recovery_anchor_operations%ROWTYPE;
v_head platform.iam_protected_authority_heads%ROWTYPE;
BEGIN
    IF p_operation_id IS NULL OR p_tenant_id IS NULL OR p_actor_user_id IS NULL OR
       p_registry_binding_id IS NULL OR p_authority_generation<2 OR
       pg_catalog.octet_length(p_protected_state_sha256)<>32 OR
       pg_catalog.octet_length(p_proposal_sha256)<>32 OR p_occurred_at IS NULL OR
       NOT platform.iam_platform_current_owner_v1(p_tenant_id,p_actor_user_id) THEN
        RAISE EXCEPTION USING ERRCODE='22023',MESSAGE='recovery anchor proposal is invalid';
    END IF;
    SELECT * INTO v_operation FROM platform.iam_recovery_anchor_operations AS operation
     WHERE operation.operation_id=p_operation_id AND operation.tenant_id=p_tenant_id
       AND operation.actor_user_id=p_actor_user_id FOR UPDATE;
    IF v_operation.state='proposal_signed' THEN
        RETURN v_operation.registry_binding_id=p_registry_binding_id AND
            v_operation.authority_generation=p_authority_generation AND
            v_operation.protected_state_sha256=p_protected_state_sha256 AND
            v_operation.proposal_sha256=p_proposal_sha256;
    END IF;
    SELECT * INTO v_head FROM platform.iam_protected_authority_heads AS head
     WHERE head.environment_id=v_operation.environment_id FOR SHARE;
    IF v_operation.state<>'authority_pending' OR v_head.registry_binding_id<>p_registry_binding_id OR
       p_authority_generation<>v_head.generation+1 OR NOT EXISTS(
        SELECT 1 FROM platform.platformauthority_operations AS authority
         WHERE authority.operation_id=p_operation_id AND authority.purpose='recovery_anchor_change'
           AND authority.binding_id=p_registry_binding_id AND authority.scope_kind='environment'
           AND authority.scope_id=v_operation.environment_id
           AND authority.generation=p_authority_generation
           AND authority.previous_head_sha256=pg_catalog.encode(v_head.ack_sha256,'hex')
           AND authority.state_sha256=pg_catalog.encode(p_protected_state_sha256,'hex')
           AND authority.proposal_sha256=pg_catalog.encode(p_proposal_sha256,'hex')
           AND authority.status='proposal_signed') THEN
        RAISE EXCEPTION USING ERRCODE='40001',MESSAGE='recovery anchor proposal conflicts';
    END IF;
    UPDATE platform.iam_recovery_anchor_operations
       SET registry_binding_id=p_registry_binding_id,authority_generation=p_authority_generation,
           protected_state_sha256=p_protected_state_sha256,proposal_sha256=p_proposal_sha256,
           state='proposal_signed',updated_at=p_occurred_at
     WHERE operation_id=p_operation_id AND state='authority_pending';
    IF NOT FOUND OR NOT platform.platformops_mark_anchor_handoff_v1(p_operation_id,p_occurred_at) THEN
        RAISE EXCEPTION USING ERRCODE='40001',MESSAGE='recovery anchor handoff conflicts';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_recovery_anchor_stage_delete_v1(
    uuid,uuid,uuid,uuid,uuid,text,bytea,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_recovery_anchor_record_proposal_v1(
    uuid,uuid,uuid,uuid,bigint,bytea,bytea,timestamptz) FROM PUBLIC;
