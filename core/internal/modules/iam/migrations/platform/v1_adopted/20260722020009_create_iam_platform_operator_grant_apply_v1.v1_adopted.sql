-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_platform_operator_grant_apply_v1(
    p_tenant_id uuid, p_operation_id uuid, p_request_sha256 bytea,
    p_actor_id uuid, p_target_user_id uuid, p_target_authorization_version bigint,
    p_expected_platform_clock bigint, p_subject_id uuid, p_role_id uuid,
    p_subject_role_binding_id uuid, p_role_permission_binding_id uuid,
    p_capability text
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
              AND receipt.mutation_kind = 'operator_grant'
              AND receipt.request_sha256 = p_request_sha256
              AND receipt.actor_user_id = p_actor_id
              AND receipt.target_user_id = p_target_user_id
              AND receipt.target_authorization_version = p_target_authorization_version
              AND permission.permission_key = p_capability
              AND subject.primary_tenant_id = p_tenant_id
        ) THEN
            RAISE EXCEPTION USING ERRCODE = 'P1401',
                MESSAGE = 'Platform operator grant idempotency conflict';
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
            MESSAGE = 'Platform capability cannot be delegated';
    END IF;

    INSERT INTO platform.iam_platform_subjects (
        id, environment_id, primary_tenant_id, user_id, state, revision,
        created_at, updated_at, retired_at
    ) VALUES (p_subject_id, v_environment_id, p_tenant_id, p_target_user_id,
        'active', 1, v_now, v_now, NULL) ON CONFLICT DO NOTHING;
    SELECT subject.id INTO v_subject_id FROM platform.iam_platform_subjects AS subject
    WHERE subject.environment_id = v_environment_id
      AND subject.primary_tenant_id = p_tenant_id
      AND subject.user_id = p_target_user_id AND subject.state = 'active' FOR SHARE;

    INSERT INTO platform.iam_platform_roles (
        id, environment_id, role_kind, role_key, state, revision,
        created_at, updated_at, retired_at
    ) VALUES (p_role_id, v_environment_id, 'platform_operator',
        'platform_operator.' || p_target_user_id::text,
        'active', 1, v_now, v_now, NULL) ON CONFLICT DO NOTHING;
    SELECT role.id INTO v_role_id FROM platform.iam_platform_roles AS role
    WHERE role.environment_id = v_environment_id
      AND role.role_key = 'platform_operator.' || p_target_user_id::text
      AND role.role_kind = 'platform_operator' AND role.state = 'active' FOR SHARE;
    IF v_subject_id IS NULL OR v_role_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P1401',
            MESSAGE = 'Platform operator identity plan conflicts with stored authority';
    END IF;

    SELECT binding.id INTO v_subject_role_id
    FROM platform.iam_platform_subject_roles AS binding
    WHERE binding.environment_id = v_environment_id
      AND binding.subject_id = v_subject_id AND binding.role_id = v_role_id
      AND binding.state = 'active' FOR SHARE;
    IF v_subject_role_id IS NULL THEN
        INSERT INTO platform.iam_platform_subject_roles (
            id, environment_id, subject_id, role_id, state, revision,
            created_at, updated_at, revoked_at
        ) VALUES (p_subject_role_binding_id, v_environment_id, v_subject_id,
            v_role_id, 'active', 1, v_now, v_now, NULL)
        RETURNING id INTO v_subject_role_id;
    END IF;
    IF EXISTS (SELECT 1 FROM platform.iam_platform_role_permissions AS binding
        WHERE binding.environment_id = v_environment_id AND binding.role_id = v_role_id
          AND binding.permission_id = v_permission_id AND binding.state = 'active') THEN
        RAISE EXCEPTION USING ERRCODE = 'P1401',
            MESSAGE = 'Platform capability is already delegated';
    END IF;
    INSERT INTO platform.iam_platform_role_permissions (
        id, environment_id, role_id, permission_id, state, revision,
        created_at, updated_at, revoked_at
    ) VALUES (p_role_permission_binding_id, v_environment_id, v_role_id,
        v_permission_id, 'active', 1, v_now, v_now, NULL)
    RETURNING id INTO v_role_permission_id;

    UPDATE platform.iam_platform_authorization_clocks
    SET authorization_clock = authorization_clock + 1, revision = revision + 1,
        updated_at = v_now
    WHERE environment_id = v_environment_id
      AND authorization_clock = p_expected_platform_clock
    RETURNING authorization_clock INTO v_clock;
    IF v_clock IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'Platform operator grant authority changed';
    END IF;
    INSERT INTO platform.iam_platform_mutation_receipts VALUES (
        p_operation_id, v_environment_id, 'operator_grant', p_request_sha256,
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
REVOKE ALL ON FUNCTION platform.iam_platform_operator_grant_apply_v1(
    uuid,uuid,bytea,uuid,uuid,bigint,bigint,uuid,uuid,uuid,uuid,text
) FROM PUBLIC;
