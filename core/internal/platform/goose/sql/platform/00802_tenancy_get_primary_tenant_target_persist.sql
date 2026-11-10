-- +goose Up
DROP FUNCTION platform.tenancy_get_primary_tenant_target_v1();

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_get_primary_tenant_target_v1()
RETURNS TABLE (
    environment_id uuid,
    deployment_mode text,
    tenant_id uuid,
    purpose text,
    slug text
)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path TO pg_catalog
AS $fn$
    WITH candidate AS (
        SELECT policy.environment_id, policy.deployment_mode, tenant.id AS tenant_id,
            fact.binding_purpose AS purpose, tenant.slug
        FROM platform.tenancy_deployment_policy AS policy
        JOIN platform.tenancy_activation_facts AS fact
            ON fact.environment_id = policy.environment_id
        JOIN platform.tenants AS tenant
            ON tenant.id = fact.tenant_id
            AND tenant.schema_name = fact.schema_name
            AND tenant.slug = fact.slug
            AND tenant.deleted_at IS NULL
            AND tenant.purged_at IS NULL
            AND tenant.status = 'ACTIVE'
        JOIN platform.tenancy_onboarding_operations AS operation
            ON operation.operation_id = fact.operation_id
            AND operation.environment_id = fact.environment_id
            AND operation.tenant_id = tenant.id
            AND operation.schema_name = tenant.schema_name
            AND operation.state = 'succeeded'
        WHERE policy.deployment_mode = 'onpremise' AND fact.binding_purpose = 'onpremise_primary'
            OR policy.deployment_mode = 'saas' AND fact.binding_purpose = 'saas_control'
    )
    SELECT candidate.environment_id, candidate.deployment_mode, candidate.tenant_id,
        candidate.purpose, candidate.slug
    FROM candidate
    WHERE (SELECT count(*) FROM candidate) = 1;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_get_primary_tenant_target_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_get_primary_tenant_target_v1() TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_get_primary_tenant_target_v1();
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_get_primary_tenant_target_v1()
RETURNS TABLE (environment_id uuid, deployment_mode text, tenant_id uuid, purpose text, slug text)
LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$
    SELECT NULL::uuid, NULL::text, NULL::uuid, NULL::text, NULL::text WHERE false;
$fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.tenancy_get_primary_tenant_target_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_get_primary_tenant_target_v1() TO harpy_api, harpy_worker;
