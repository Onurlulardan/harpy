-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_assert_saas_customer_offboard_v1(
    p_environment_id uuid,
    p_tenant_id uuid
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN platform.tenancy_assert_saas_customer_target_v1(
        p_environment_id,
        p_tenant_id
    );
END;
$function$;
CREATE FUNCTION platform.tenancy_list_saas_customer_tenants_v1(
    p_environment_id uuid
)
RETURNS TABLE (tenant_id uuid)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
        OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
        OR p_environment_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'SaaS customer list query denied';
    END IF;
    PERFORM 1
    FROM platform.tenancy_deployment_policy AS policy
    WHERE policy.environment_id = p_environment_id
      AND policy.deployment_mode = 'saas';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'SaaS deployment policy is required';
    END IF;
    RETURN QUERY
    SELECT binding.tenant_id
    FROM platform.tenancy_tenant_bindings AS binding
    WHERE binding.environment_id = p_environment_id
      AND binding.purpose = 'saas_customer'
    ORDER BY binding.tenant_id;
END;
$function$;
CREATE FUNCTION platform.tenancy_get_primary_tenant_target_v1()
RETURNS TABLE (
    environment_id uuid,
    deployment_mode text,
    tenant_id uuid,
    purpose text,
    slug text
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
        OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Primary tenant target query denied';
    END IF;
    RETURN QUERY
    SELECT policy.environment_id,
           policy.deployment_mode,
           binding.tenant_id,
           binding.purpose,
           tenant.slug::text
    FROM platform.tenancy_deployment_policy AS policy
    JOIN platform.tenancy_tenant_bindings AS binding
      ON binding.environment_id = policy.environment_id
     AND binding.purpose = CASE policy.deployment_mode
         WHEN 'onpremise' THEN 'onpremise_primary'
         WHEN 'saas' THEN 'saas_control'
     END
    JOIN platform.tenants AS tenant ON tenant.id = binding.tenant_id
    WHERE policy.singleton
    FOR SHARE OF policy, binding, tenant;
END;
$function$;
-- +goose StatementEnd
