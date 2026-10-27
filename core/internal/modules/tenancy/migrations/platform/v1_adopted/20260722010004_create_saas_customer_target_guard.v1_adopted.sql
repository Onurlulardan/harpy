-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_assert_saas_customer_target_v1(
    p_environment_id uuid,
    p_tenant_id uuid
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
        OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
        OR p_environment_id IS NULL OR p_tenant_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'SaaS customer target query denied';
    END IF;
    PERFORM 1
    FROM platform.tenancy_deployment_policy AS policy
    JOIN platform.tenancy_tenant_bindings AS binding
      ON binding.environment_id = policy.environment_id
    WHERE policy.environment_id = p_environment_id
      AND policy.deployment_mode = 'saas'
      AND binding.tenant_id = p_tenant_id
      AND binding.purpose = 'saas_customer';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'SaaS customer target is required';
    END IF;
    RETURN p_tenant_id;
END;
$function$;
-- +goose StatementEnd
