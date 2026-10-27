-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_platform_authorize_decide_v1(
    p_tenant_id uuid, p_user_id uuid, p_expected_platform_clock bigint,
    p_capability text
)
RETURNS TABLE (
    allowed boolean, deployment_mode text, subject_id uuid,
    platform_clock bigint, reason text
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_environment_id uuid;
    v_mode text;
    v_purpose text;
    v_subject_id uuid;
    v_clock bigint;
    v_permission_id uuid;
    v_allowed boolean := false;
BEGIN
    SELECT context.environment_id, context.deployment_mode,
           context.tenant_purpose, context.subject_id, context.platform_clock
    INTO v_environment_id, v_mode, v_purpose, v_subject_id, v_clock
    FROM platform.iam_platform_session_context_read_v1(
        p_tenant_id, p_user_id
    ) AS context;
    IF v_environment_id IS NULL OR v_purpose = 'saas_customer' OR
       v_clock <> p_expected_platform_clock THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'Platform authorization context is stale or unavailable';
    END IF;
    SELECT permission.id INTO v_permission_id
    FROM platform.iam_platform_permissions AS permission
    WHERE permission.permission_key = p_capability
      AND permission.match_kind = 'exact' AND permission.state = 'active'
      AND permission.deployment_mode IN ('both', v_mode);
    IF v_subject_id IS NOT NULL AND v_permission_id IS NOT NULL THEN
        SELECT EXISTS (
            SELECT 1
            FROM platform.iam_platform_subject_roles AS subject_binding
            JOIN platform.iam_platform_roles AS role
              ON role.environment_id = subject_binding.environment_id
             AND role.id = subject_binding.role_id AND role.state = 'active'
            JOIN platform.iam_platform_role_permissions AS role_permission
              ON role_permission.environment_id = role.environment_id
             AND role_permission.role_id = role.id AND role_permission.state = 'active'
            JOIN platform.iam_platform_permissions AS granted
              ON granted.id = role_permission.permission_id AND granted.state = 'active'
            LEFT JOIN platform.iam_platform_owner_policies AS owner_policy
              ON owner_policy.environment_id = subject_binding.environment_id
             AND owner_policy.owner_binding_id = subject_binding.id
            WHERE subject_binding.environment_id = v_environment_id
              AND subject_binding.subject_id = v_subject_id
              AND subject_binding.state = 'active'
              AND ((role.role_kind = 'instance_owner'
                    AND owner_policy.owner_binding_id IS NOT NULL
                    AND granted.match_kind = 'environment_all')
                OR (role.role_kind = 'platform_operator'
                    AND granted.id = v_permission_id
                    AND granted.match_kind = 'exact' AND granted.delegable))
        ) INTO v_allowed;
    END IF;
    allowed := v_allowed;
    deployment_mode := v_mode;
    subject_id := v_subject_id;
    platform_clock := v_clock;
    reason := CASE WHEN v_allowed THEN 'granted' ELSE 'not_granted' END;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.iam_platform_authorize_decide_v1(
    uuid,uuid,bigint,text
) FROM PUBLIC;
