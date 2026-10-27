-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_platform_operator_revoke_apply_v1(
    p_tenant_id uuid, p_operation_id uuid, p_request_sha256 bytea,
    p_actor_id uuid, p_target_user_id uuid, p_target_authorization_version bigint,
    p_expected_platform_clock bigint, p_capability text
)
RETURNS TABLE (
    target_user_id uuid, subject_id uuid, role_id uuid,
    subject_role_binding_id uuid, role_permission_binding_id uuid,
    target_authorization_version bigint, platform_clock bigint, replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_environment_id uuid;
    v_mode text;
    v_owner_subject_id uuid;
    v_prior_clock bigint;
    v_permission_id uuid;
    v_subject_id uuid;
    v_role_id uuid;
    v_subject_role_id uuid;
    v_role_permission_id uuid;
    v_clock bigint;
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF EXISTS (SELECT 1 FROM platform.iam_platform_mutation_receipts AS receipt
               WHERE receipt.operation_id = p_operation_id) THEN
        IF NOT EXISTS (
            SELECT 1 FROM platform.iam_platform_mutation_receipts AS receipt
            JOIN platform.iam_platform_permissions AS permission
              ON permission.id = receipt.permission_id
            JOIN platform.iam_platform_subjects AS subject
              ON subject.environment_id = receipt.environment_id
             AND subject.id = receipt.subject_id
            WHERE receipt.operation_id = p_operation_id
              AND receipt.mutation_kind = 'operator_revoke'
              AND receipt.request_sha256 = p_request_sha256
              AND receipt.actor_user_id = p_actor_id
              AND receipt.target_user_id = p_target_user_id
              AND receipt.target_authorization_version = p_target_authorization_version
              AND permission.permission_key = p_capability
              AND subject.primary_tenant_id = p_tenant_id
        ) THEN
            RAISE EXCEPTION USING ERRCODE = 'P1401',
                MESSAGE = 'Platform operator revoke idempotency conflict';
        END IF;
        RETURN QUERY SELECT receipt.target_user_id, receipt.subject_id,
            receipt.role_id, receipt.subject_role_binding_id,
            receipt.role_permission_binding_id,
            receipt.target_authorization_version, receipt.platform_clock, true
        FROM platform.iam_platform_mutation_receipts AS receipt
        WHERE receipt.operation_id = p_operation_id;
        RETURN;
    END IF;

    SELECT context.environment_id, context.deployment_mode,
           context.owner_subject_id, context.platform_clock
    INTO v_environment_id, v_mode, v_owner_subject_id, v_prior_clock
    FROM platform.iam_platform_owner_context_lock_v1(
        p_tenant_id, p_actor_id, p_expected_platform_clock
    ) AS context;
    IF v_environment_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Current actor is not the Instance Owner';
    END IF;
    SELECT permission.id INTO v_permission_id
    FROM platform.iam_platform_permissions AS permission
    WHERE permission.permission_key = p_capability AND permission.match_kind = 'exact'
      AND permission.state = 'active' AND permission.delegable
      AND NOT permission.owner_only
      AND permission.deployment_mode IN ('both', v_mode) FOR SHARE;
    IF v_permission_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Platform capability cannot be revoked';
    END IF;

    SELECT subject.id, role.id, subject_binding.id, permission_binding.id
    INTO v_subject_id, v_role_id, v_subject_role_id, v_role_permission_id
    FROM platform.iam_platform_subjects AS subject
    JOIN platform.iam_platform_subject_roles AS subject_binding
      ON subject_binding.environment_id = subject.environment_id
     AND subject_binding.subject_id = subject.id AND subject_binding.state = 'active'
    JOIN platform.iam_platform_roles AS role
      ON role.environment_id = subject_binding.environment_id
     AND role.id = subject_binding.role_id
     AND role.role_kind = 'platform_operator' AND role.state = 'active'
    JOIN platform.iam_platform_role_permissions AS permission_binding
      ON permission_binding.environment_id = role.environment_id
     AND permission_binding.role_id = role.id
     AND permission_binding.permission_id = v_permission_id
     AND permission_binding.state = 'active'
    WHERE subject.environment_id = v_environment_id
      AND subject.primary_tenant_id = p_tenant_id
      AND subject.user_id = p_target_user_id AND subject.state = 'active'
    FOR UPDATE OF permission_binding
    FOR SHARE OF subject, subject_binding, role;
    IF v_role_permission_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P1404',
            MESSAGE = 'Platform capability delegation was not found';
    END IF;
    UPDATE platform.iam_platform_role_permissions
    SET state = 'revoked', revision = revision + 1,
        updated_at = v_now, revoked_at = v_now
    WHERE environment_id = v_environment_id
      AND id = v_role_permission_id AND state = 'active';
    UPDATE platform.iam_platform_authorization_clocks
    SET authorization_clock = authorization_clock + 1, revision = revision + 1,
        updated_at = v_now
    WHERE environment_id = v_environment_id
      AND authorization_clock = p_expected_platform_clock
    RETURNING authorization_clock INTO v_clock;
    IF v_clock IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'Platform operator revoke authority changed';
    END IF;
    INSERT INTO platform.iam_platform_mutation_receipts VALUES (
        p_operation_id, v_environment_id, 'operator_revoke', p_request_sha256,
        p_actor_id, p_target_user_id, v_subject_id, v_role_id, v_subject_role_id,
        v_role_permission_id, v_permission_id, p_target_authorization_version,
        v_clock, v_now
    );
    RETURN QUERY SELECT p_target_user_id, v_subject_id, v_role_id,
        v_subject_role_id, v_role_permission_id,
        p_target_authorization_version, v_clock, false;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.iam_platform_operator_revoke_apply_v1(
    uuid,uuid,bytea,uuid,uuid,bigint,bigint,text
) FROM PUBLIC;
