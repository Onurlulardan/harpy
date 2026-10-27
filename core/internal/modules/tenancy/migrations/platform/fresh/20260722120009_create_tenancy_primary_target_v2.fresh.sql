-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_get_primary_tenant_target_v2()
RETURNS TABLE (
    environment_id uuid,
    deployment_mode text,
    tenant_id uuid,
    purpose text,
    slug text,
    schema_name name,
    tenant_set_sha256 bytea
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
    WITH tenant_set AS (
        SELECT policy.environment_id,
               pg_catalog.sha256(pg_catalog.convert_to(
                   'tenant-set.v1' || COALESCE(pg_catalog.string_agg(
                       '|' || binding.tenant_id::text || '|' ||
                       tenant.schema_name::text,
                       '' ORDER BY binding.tenant_id::text
                   ), ''),
                   'UTF8'
               )) AS digest
        FROM platform.tenancy_deployment_policy AS policy
        JOIN platform.tenancy_tenant_bindings AS binding
          ON binding.environment_id = policy.environment_id
        JOIN platform.tenants AS tenant ON tenant.id = binding.tenant_id
        WHERE policy.singleton
        GROUP BY policy.environment_id
    )
    SELECT policy.environment_id,
           policy.deployment_mode,
           binding.tenant_id,
           binding.purpose,
           tenant.slug::text,
           tenant.schema_name,
           tenant_set.digest
    FROM platform.tenancy_deployment_policy AS policy
    JOIN platform.tenancy_tenant_bindings AS binding
      ON binding.environment_id = policy.environment_id
     AND binding.purpose = CASE policy.deployment_mode
         WHEN 'onpremise' THEN 'onpremise_primary'
         WHEN 'saas' THEN 'saas_control'
     END
    JOIN platform.tenants AS tenant ON tenant.id = binding.tenant_id
    JOIN tenant_set ON tenant_set.environment_id = policy.environment_id
    WHERE policy.singleton
    FOR SHARE OF policy, binding, tenant;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_get_primary_tenant_target_v2()
    FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_get_primary_tenant_target_v2()
    TO harpy_runtime;
