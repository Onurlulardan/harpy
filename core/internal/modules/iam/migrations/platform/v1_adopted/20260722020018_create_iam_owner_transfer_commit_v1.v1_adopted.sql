-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_owner_transfer_commit_v1(
    p_tenant_id uuid, p_operation_id uuid, p_request_sha256 bytea,
    p_actor_id uuid, p_expected_platform_clock bigint
)
RETURNS TABLE (
    operation_id uuid, predecessor_user_id uuid, successor_user_id uuid,
    predecessor_token_version bigint, predecessor_authorization_version bigint,
    successor_token_version bigint, successor_authorization_version bigint,
    predecessor_expected_revision bigint, successor_expected_revision bigint,
    owner_policy_revision bigint, platform_clock bigint, head_generation bigint,
    replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_environment_id uuid; v_owner_subject_id uuid; v_clock bigint;
    v_successor_user_id uuid; v_successor_subject_plan uuid; v_successor_binding_plan uuid;
    v_successor_subject_id uuid; v_successor_binding_id uuid; v_owner_role_id uuid;
    v_predecessor_binding_id uuid; v_policy_revision bigint; v_generation bigint;
    v_previous_head bytea; v_state_sha bytea; v_ack_sha bytea;
    v_actor_token bigint; v_actor_auth bigint; v_successor_token bigint; v_successor_auth bigint;
    v_actor_revision bigint; v_successor_revision bigint;
    v_status text; v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF p_operation_id IS NULL OR pg_catalog.octet_length(p_request_sha256) <> 32 OR
       p_actor_id IS NULL OR p_expected_platform_clock < 1 THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'Owner transfer commit is invalid';
    END IF;
    SELECT operation.status, operation.successor_user_id,
           operation.successor_subject_id, operation.successor_binding_id,
           operation.actor_expected_token_version, operation.actor_expected_authorization_version,
           operation.successor_expected_token_version, operation.successor_expected_authorization_version,
           operation.actor_expected_revision, operation.successor_expected_revision,
           operation.generation, operation.previous_head_sha256,
           operation.state_sha256, operation.ack_sha256,
           operation.owner_policy_revision, operation.platform_clock
    INTO v_status, v_successor_user_id, v_successor_subject_plan, v_successor_binding_plan,
         v_actor_token, v_actor_auth, v_successor_token, v_successor_auth,
         v_actor_revision, v_successor_revision,
         v_generation, v_previous_head, v_state_sha, v_ack_sha, v_policy_revision, v_clock
    FROM platform.iam_platform_owner_transfer_operations AS operation
    WHERE operation.operation_id = p_operation_id AND operation.actor_user_id = p_actor_id
      AND operation.request_sha256 = p_request_sha256 FOR UPDATE;
    IF v_status = 'completed' THEN
        RETURN QUERY SELECT p_operation_id,p_actor_id,v_successor_user_id,
            v_actor_token + 1,v_actor_auth + 1,v_successor_token + 1,v_successor_auth + 1,
            v_actor_revision,v_successor_revision,
            v_policy_revision,v_clock,v_generation,true;
        RETURN;
    END IF;
    IF v_status <> 'ack_verified' THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'Owner transfer ack is not current';
    END IF;
    SELECT context.environment_id, context.owner_subject_id, context.platform_clock
    INTO v_environment_id, v_owner_subject_id, v_clock
    FROM platform.iam_platform_owner_context_lock_v1(
        p_tenant_id, p_actor_id, p_expected_platform_clock
    ) AS context;
    SELECT policy.owner_binding_id, policy.revision, binding.role_id
    INTO v_predecessor_binding_id, v_policy_revision, v_owner_role_id
    FROM platform.iam_platform_owner_policies AS policy
    JOIN platform.iam_platform_subject_roles AS binding
      ON binding.environment_id = policy.environment_id AND binding.id = policy.owner_binding_id
    WHERE policy.environment_id = v_environment_id FOR UPDATE OF policy, binding;
    IF v_environment_id IS NULL OR v_policy_revision IS NULL OR
       v_generation IS NULL OR v_clock <> p_expected_platform_clock THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'Owner transfer authority changed';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM platform.iam_platform_owner_transfer_operations AS operation
        WHERE operation.operation_id = p_operation_id AND operation.environment_id = v_environment_id
          AND operation.expected_owner_policy_revision = v_policy_revision
          AND operation.expected_platform_clock = p_expected_platform_clock) THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'Owner transfer policy changed';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM platform.iam_protected_authority_heads AS head
        WHERE head.environment_id = v_environment_id AND head.generation = v_generation - 1
          AND head.state_sha256 = v_previous_head FOR UPDATE) THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'Owner transfer head changed';
    END IF;
    INSERT INTO platform.iam_platform_subjects (
        id,environment_id,primary_tenant_id,user_id,state,revision,created_at,updated_at,retired_at
    ) VALUES (v_successor_subject_plan,v_environment_id,p_tenant_id,v_successor_user_id,
        'active',1,v_now,v_now,NULL) ON CONFLICT DO NOTHING;
    SELECT subject.id INTO v_successor_subject_id
    FROM platform.iam_platform_subjects AS subject
    WHERE subject.environment_id = v_environment_id AND subject.primary_tenant_id = p_tenant_id
      AND subject.user_id = v_successor_user_id AND subject.state = 'active' FOR SHARE;
    IF v_successor_subject_id <> v_successor_subject_plan THEN
        RAISE EXCEPTION USING ERRCODE = 'P1401', MESSAGE = 'Owner successor subject conflicts';
    END IF;
    INSERT INTO platform.iam_platform_subject_roles (
        id,environment_id,subject_id,role_id,state,revision,created_at,updated_at,revoked_at
    ) VALUES (v_successor_binding_plan,v_environment_id,v_successor_subject_id,
        v_owner_role_id,'active',1,v_now,v_now,NULL);
    v_successor_binding_id := v_successor_binding_plan;
    UPDATE platform.iam_platform_owner_policies
    SET owner_binding_id = v_successor_binding_id, revision = revision + 1, updated_at = v_now
    WHERE environment_id = v_environment_id AND owner_binding_id = v_predecessor_binding_id
      AND revision = v_policy_revision RETURNING revision INTO v_policy_revision;
    IF v_policy_revision IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'Owner transfer policy CAS failed';
    END IF;
    UPDATE platform.iam_platform_subject_roles
    SET state = 'revoked', revision = revision + 1, updated_at = v_now, revoked_at = v_now
    WHERE environment_id = v_environment_id AND id = v_predecessor_binding_id AND state = 'active';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'Owner predecessor changed';
    END IF;
    UPDATE platform.iam_platform_authorization_clocks
    SET authorization_clock = authorization_clock + 1, revision = revision + 1, updated_at = v_now
    WHERE environment_id = v_environment_id AND authorization_clock = p_expected_platform_clock
    RETURNING authorization_clock INTO v_clock;
    UPDATE platform.iam_protected_authority_heads
    SET generation = v_generation, state_sha256 = v_state_sha, ack_sha256 = v_ack_sha,
        revision = revision + 1, updated_at = v_now
    WHERE environment_id = v_environment_id AND generation = v_generation - 1
      AND state_sha256 = v_previous_head;
    IF NOT FOUND OR v_clock IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'Owner transfer commit CAS failed';
    END IF;
    UPDATE platform.iam_platform_owner_transfer_operations AS operation
    SET status = 'completed', predecessor_token_version = v_actor_token + 1,
        predecessor_authorization_version = v_actor_auth + 1,
        successor_token_version = v_successor_token + 1,
        successor_authorization_version = v_successor_auth + 1,
        owner_policy_revision = v_policy_revision, platform_clock = v_clock,
        revision = operation.revision + 1, updated_at = v_now, completed_at = v_now
    WHERE operation.operation_id = p_operation_id AND operation.status = 'ack_verified';
    RETURN QUERY SELECT p_operation_id,p_actor_id,v_successor_user_id,
        v_actor_token + 1,v_actor_auth + 1,v_successor_token + 1,v_successor_auth + 1,
        v_actor_revision,v_successor_revision,
        v_policy_revision,v_clock,v_generation,false;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.iam_owner_transfer_commit_v1(
    uuid,uuid,bytea,uuid,bigint
) FROM PUBLIC;
