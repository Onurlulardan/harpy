-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_platform_session_context_read_v1(
    p_tenant_id uuid, p_user_id uuid
)
RETURNS TABLE (
    environment_id uuid, deployment_mode text, tenant_purpose text,
    subject_id uuid, platform_clock bigint
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
BEGIN
    RETURN QUERY
    SELECT binding.environment_id, deployment.deployment_mode, binding.purpose,
           CASE WHEN binding.purpose = 'saas_customer' THEN NULL ELSE subject.id END,
           CASE WHEN binding.purpose = 'saas_customer' THEN 0
                ELSE clock.authorization_clock END
    FROM platform.tenancy_tenant_bindings AS binding
    JOIN platform.tenancy_deployment_policy AS deployment
      ON deployment.environment_id = binding.environment_id
    LEFT JOIN platform.iam_platform_authorization_clocks AS clock
      ON clock.environment_id = binding.environment_id
    LEFT JOIN platform.iam_platform_subjects AS subject
      ON subject.environment_id = binding.environment_id
     AND subject.primary_tenant_id = p_tenant_id
     AND subject.user_id = p_user_id AND subject.state = 'active'
    WHERE binding.tenant_id = p_tenant_id
      AND (binding.purpose = 'saas_customer' OR clock.authorization_clock > 0);
END;
$function$;

CREATE FUNCTION platform.iam_platform_current_owner_v1(
    p_tenant_id uuid, p_user_id uuid
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM platform.iam_platform_owner_policies AS policy
        JOIN platform.iam_platform_subject_roles AS owner_binding
          ON owner_binding.environment_id = policy.environment_id
         AND owner_binding.id = policy.owner_binding_id AND owner_binding.state = 'active'
        JOIN platform.iam_platform_subjects AS subject
          ON subject.environment_id = owner_binding.environment_id
         AND subject.id = owner_binding.subject_id AND subject.state = 'active'
        WHERE subject.primary_tenant_id = p_tenant_id AND subject.user_id = p_user_id
    );
END;
$function$;

CREATE FUNCTION platform.iam_platform_owner_context_lock_v1(
    p_tenant_id uuid, p_user_id uuid, p_expected_platform_clock bigint
)
RETURNS TABLE (
    environment_id uuid, deployment_mode text, owner_subject_id uuid,
    platform_clock bigint
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
BEGIN
    RETURN QUERY
    SELECT policy.environment_id, deployment.deployment_mode,
           subject.id, clock.authorization_clock
    FROM platform.tenancy_tenant_bindings AS tenant_binding
    JOIN platform.tenancy_deployment_policy AS deployment
      ON deployment.environment_id = tenant_binding.environment_id
    JOIN platform.iam_platform_owner_policies AS policy
      ON policy.environment_id = deployment.environment_id
    JOIN platform.iam_platform_subject_roles AS owner_binding
      ON owner_binding.environment_id = policy.environment_id
     AND owner_binding.id = policy.owner_binding_id AND owner_binding.state = 'active'
    JOIN platform.iam_platform_roles AS owner_role
      ON owner_role.environment_id = owner_binding.environment_id
     AND owner_role.id = owner_binding.role_id
     AND owner_role.role_kind = 'instance_owner' AND owner_role.state = 'active'
    JOIN platform.iam_platform_subjects AS subject
      ON subject.environment_id = owner_binding.environment_id
     AND subject.id = owner_binding.subject_id AND subject.state = 'active'
     AND subject.primary_tenant_id = p_tenant_id AND subject.user_id = p_user_id
    JOIN platform.iam_platform_authorization_clocks AS clock
      ON clock.environment_id = policy.environment_id
    WHERE tenant_binding.tenant_id = p_tenant_id
      AND tenant_binding.purpose IN ('onpremise_primary', 'saas_control')
      AND clock.authorization_clock = p_expected_platform_clock
    FOR UPDATE OF clock
    FOR SHARE OF tenant_binding, deployment, policy, owner_binding, owner_role, subject;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.iam_platform_session_context_read_v1(uuid,uuid)
FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_platform_current_owner_v1(uuid,uuid)
FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_platform_owner_context_lock_v1(uuid,uuid,bigint)
FROM PUBLIC;
