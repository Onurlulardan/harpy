-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_owner_transfer_prepare_v1(
    p_tenant_id uuid, p_operation_id uuid, p_request_sha256 bytea,
    p_actor_id uuid, p_actor_revision bigint, p_actor_token_version bigint,
    p_actor_authorization_version bigint, p_successor_id uuid,
    p_successor_revision bigint, p_successor_token_version bigint,
    p_successor_authorization_version bigint, p_expected_platform_clock bigint,
    p_successor_subject_id uuid, p_successor_binding_id uuid,
    p_registry_binding_id uuid, p_expected_policy_revision bigint,
    p_expected_head_generation bigint, p_expected_head_sha256 bytea,
    p_state_sha256 bytea, p_proposal_sha256 bytea, p_proposal_signer_kid text,
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
    v_policy_revision bigint; v_head_generation bigint; v_head_sha bytea;
    v_existing_subject_id uuid; v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF p_operation_id IS NULL OR pg_catalog.octet_length(p_request_sha256) <> 32 OR
       p_actor_id IS NULL OR p_successor_id IS NULL OR p_actor_id = p_successor_id OR
       p_actor_revision < 1 OR p_actor_token_version < 1 OR p_actor_authorization_version < 1 OR
       p_successor_revision < 1 OR p_successor_token_version < 1 OR
       p_successor_authorization_version < 1 OR p_successor_subject_id IS NULL OR
       p_successor_binding_id IS NULL OR p_registry_binding_id IS NULL OR
       p_expected_policy_revision < 1 OR p_expected_head_generation < 1 OR
       pg_catalog.octet_length(p_expected_head_sha256) <> 32 OR
       pg_catalog.octet_length(p_state_sha256) <> 32 OR p_state_sha256 = p_expected_head_sha256 OR
       pg_catalog.octet_length(p_proposal_sha256) <> 32 OR
       p_proposal_signer_kid !~ '^[a-z0-9][a-z0-9._-]{0,127}$' OR
       pg_catalog.octet_length(p_proposal_signature_sha256) <> 32 OR
       p_proposal_expires_at < v_now OR p_proposal_expires_at > v_now + INTERVAL '7 days' THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'Owner transfer proposal is invalid';
    END IF;
    IF EXISTS (SELECT 1 FROM platform.iam_platform_owner_transfer_operations AS operation
               WHERE operation.operation_id = p_operation_id) THEN
        IF NOT EXISTS (SELECT 1 FROM platform.iam_platform_owner_transfer_operations AS operation
            WHERE operation.operation_id = p_operation_id AND operation.request_sha256 = p_request_sha256
              AND operation.actor_user_id = p_actor_id AND operation.successor_user_id = p_successor_id
              AND operation.proposal_sha256 = p_proposal_sha256 AND operation.status = 'prepared') THEN
            RAISE EXCEPTION USING ERRCODE = 'P1401', MESSAGE = 'Owner transfer idempotency conflict';
        END IF;
        RETURN QUERY SELECT operation.operation_id, operation.registry_binding_id,
            operation.scope_id, operation.generation, operation.previous_head_sha256,
            operation.state_sha256, operation.proposal_sha256, operation.status, true
        FROM platform.iam_platform_owner_transfer_operations AS operation
        WHERE operation.operation_id = p_operation_id;
        RETURN;
    END IF;
    SELECT context.environment_id, context.owner_subject_id, context.platform_clock
    INTO v_environment_id, v_owner_subject_id, v_clock
    FROM platform.iam_platform_owner_context_lock_v1(
        p_tenant_id, p_actor_id, p_expected_platform_clock
    ) AS context;
    IF v_environment_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Current actor is not the Instance Owner';
    END IF;
    SELECT policy.revision INTO v_policy_revision
    FROM platform.iam_platform_owner_policies AS policy
    WHERE policy.environment_id = v_environment_id FOR UPDATE;
    SELECT head.generation, head.state_sha256 INTO v_head_generation, v_head_sha
    FROM platform.iam_protected_authority_heads AS head
    WHERE head.environment_id = v_environment_id AND head.registry_binding_id = p_registry_binding_id
      AND head.scope_kind = 'environment' AND head.scope_id = v_environment_id FOR UPDATE;
    SELECT subject.id INTO v_existing_subject_id
    FROM platform.iam_platform_subjects AS subject
    WHERE subject.environment_id = v_environment_id AND subject.primary_tenant_id = p_tenant_id
      AND subject.user_id = p_successor_id AND subject.state = 'active' FOR SHARE;
    IF v_policy_revision <> p_expected_policy_revision OR v_head_generation <> p_expected_head_generation OR
       v_head_sha <> p_expected_head_sha256 OR
       (v_existing_subject_id IS NOT NULL AND v_existing_subject_id <> p_successor_subject_id) THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'Owner transfer authority head is stale';
    END IF;
    INSERT INTO platform.iam_platform_owner_transfer_operations (
        operation_id,environment_id,request_sha256,actor_user_id,successor_user_id,
        actor_expected_revision,actor_expected_token_version,actor_expected_authorization_version,
        successor_expected_revision,successor_expected_token_version,successor_expected_authorization_version,
        successor_subject_id,successor_binding_id,expected_owner_policy_revision,expected_platform_clock,
        registry_binding_id,scope_kind,scope_id,generation,previous_head_sha256,state_sha256,
        proposal_sha256,proposal_signer_kid,proposal_signature_sha256,proposal_expires_at,
        status,revision,created_at,updated_at
    ) VALUES (p_operation_id,v_environment_id,p_request_sha256,p_actor_id,p_successor_id,
        p_actor_revision,p_actor_token_version,p_actor_authorization_version,
        p_successor_revision,p_successor_token_version,p_successor_authorization_version,
        p_successor_subject_id,p_successor_binding_id,p_expected_policy_revision,p_expected_platform_clock,
        p_registry_binding_id,'environment',v_environment_id,p_expected_head_generation + 1,
        p_expected_head_sha256,p_state_sha256,p_proposal_sha256,p_proposal_signer_kid,
        p_proposal_signature_sha256,p_proposal_expires_at,'prepared',1,v_now,v_now);
    RETURN QUERY SELECT p_operation_id,p_registry_binding_id,v_environment_id,
        p_expected_head_generation + 1,p_expected_head_sha256,p_state_sha256,
        p_proposal_sha256,'prepared'::text,false;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.iam_owner_transfer_prepare_v1(
    uuid,uuid,bytea,uuid,bigint,bigint,bigint,uuid,bigint,bigint,bigint,bigint,
    uuid,uuid,uuid,bigint,bigint,bytea,bytea,bytea,text,bytea,timestamptz
) FROM PUBLIC;
