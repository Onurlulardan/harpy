-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_recovery_anchor_commit_v1(
    p_operation_id uuid,p_tenant_id uuid,p_actor_user_id uuid,p_occurred_at timestamptz
)
RETURNS TABLE(operation_id uuid,anchor_id uuid,kind text,authority_generation bigint,replayed boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_operation platform.iam_recovery_anchor_operations%ROWTYPE;
v_authority platform.platformauthority_operations%ROWTYPE;
v_head platform.iam_protected_authority_heads%ROWTYPE;
v_ack bytea;
v_clock bigint;
BEGIN
    IF p_operation_id IS NULL OR p_tenant_id IS NULL OR p_actor_user_id IS NULL OR
       p_occurred_at IS NULL OR
       NOT platform.iam_platform_current_owner_v1(p_tenant_id,p_actor_user_id) THEN
        RAISE EXCEPTION USING ERRCODE='22023',MESSAGE='recovery anchor commit is invalid';
    END IF;
    SELECT * INTO v_operation FROM platform.iam_recovery_anchor_operations AS operation
     WHERE operation.operation_id=p_operation_id AND operation.tenant_id=p_tenant_id
       AND operation.actor_user_id=p_actor_user_id FOR UPDATE;
    IF v_operation.state='completed' THEN
        RETURN QUERY SELECT v_operation.operation_id,v_operation.anchor_id,v_operation.kind,
            v_operation.authority_generation,true;
        RETURN;
    END IF;
    SELECT * INTO v_authority FROM platform.platformauthority_operations AS authority
     WHERE authority.operation_id=p_operation_id FOR SHARE;
    SELECT * INTO v_head FROM platform.iam_protected_authority_heads AS head
     WHERE head.environment_id=v_operation.environment_id FOR UPDATE;
    IF v_operation.state<>'proposal_signed' OR v_authority.status<>'completed' OR
       v_authority.purpose<>'recovery_anchor_change' OR
       v_authority.binding_id<>v_operation.registry_binding_id OR
       v_authority.scope_kind<>'environment' OR
       v_authority.scope_id<>v_operation.environment_id OR
       v_authority.generation<>v_operation.authority_generation OR
       v_authority.state_sha256<>pg_catalog.encode(v_operation.protected_state_sha256,'hex') OR
       v_authority.proposal_sha256<>pg_catalog.encode(v_operation.proposal_sha256,'hex') OR
       v_authority.previous_head_sha256<>pg_catalog.encode(v_head.ack_sha256,'hex') OR
       v_head.registry_binding_id<>v_operation.registry_binding_id OR
       v_head.generation+1<>v_operation.authority_generation OR
       v_authority.ack_sha256 IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='40001',MESSAGE='recovery anchor authority is not complete';
    END IF;
    v_ack:=pg_catalog.decode(v_authority.ack_sha256,'hex');
    IF pg_catalog.octet_length(v_ack)<>32 THEN
        RAISE EXCEPTION USING ERRCODE='40001',MESSAGE='recovery anchor ack is invalid';
    END IF;
    IF v_operation.kind='anchor_register' THEN
        UPDATE platform.iam_recovery_anchors
           SET state='active',authority_generation=v_operation.authority_generation,
               protected_state_sha256=v_operation.protected_state_sha256,
               authority_ack_sha256=v_ack,revision=revision+1,updated_at=p_occurred_at
         WHERE anchor_id=v_operation.anchor_id AND environment_id=v_operation.environment_id
           AND state='pending';
    ELSE
        IF (SELECT COUNT(*) FROM platform.iam_recovery_anchors AS anchor
             WHERE anchor.environment_id=v_operation.environment_id AND anchor.state='active')<=1 THEN
            RAISE EXCEPTION USING ERRCODE='40001',MESSAGE='last recovery anchor cannot be deleted';
        END IF;
        UPDATE platform.iam_recovery_anchors
           SET state='deleted',authority_generation=v_operation.authority_generation,
               protected_state_sha256=v_operation.protected_state_sha256,
               authority_ack_sha256=v_ack,revision=revision+1,updated_at=p_occurred_at,
               deleted_at=p_occurred_at
         WHERE anchor_id=v_operation.anchor_id AND environment_id=v_operation.environment_id
           AND state='active';
    END IF;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',MESSAGE='recovery anchor state changed';
    END IF;
    UPDATE platform.iam_protected_authority_heads
       SET generation=v_operation.authority_generation,
           state_sha256=v_operation.protected_state_sha256,ack_sha256=v_ack,
           revision=revision+1,updated_at=p_occurred_at
     WHERE environment_id=v_operation.environment_id
       AND generation=v_operation.authority_generation-1
       AND ack_sha256=pg_catalog.decode(v_authority.previous_head_sha256,'hex');
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE='40001',MESSAGE='recovery anchor head changed'; END IF;
    UPDATE platform.iam_platform_authorization_clocks
       SET authorization_clock=authorization_clock+1,revision=revision+1,updated_at=p_occurred_at
     WHERE environment_id=v_operation.environment_id
     RETURNING authorization_clock INTO v_clock;
    IF v_clock IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='40001',MESSAGE='recovery anchor clock changed';
    END IF;
    PERFORM platform.platformops_stale_recovery_sets_for_authority_v1(
        v_operation.environment_id,v_operation.authority_generation,
        v_operation.protected_state_sha256,p_occurred_at);
    IF NOT platform.platformops_complete_anchor_operation_v1(
        p_operation_id,v_ack,p_occurred_at) THEN
        RAISE EXCEPTION USING ERRCODE='40001',MESSAGE='recovery operation completion conflicts';
    END IF;
    UPDATE platform.iam_recovery_anchor_operations
       SET state='completed',completed_at=p_occurred_at,updated_at=p_occurred_at
     WHERE operation_id=p_operation_id AND state='proposal_signed';
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE='40001',MESSAGE='recovery anchor operation changed'; END IF;
    RETURN QUERY SELECT v_operation.operation_id,v_operation.anchor_id,v_operation.kind,
        v_operation.authority_generation,false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_recovery_anchor_commit_v1(
    uuid,uuid,uuid,timestamptz) FROM PUBLIC;
