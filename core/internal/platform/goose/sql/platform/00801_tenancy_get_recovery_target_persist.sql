-- +goose Up
DROP FUNCTION platform.tenancy_get_recovery_target_v1(text);

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_get_recovery_target_v1(p_purpose text)
RETURNS TABLE (
    environment_id uuid,
    tenant_id uuid,
    tenant_slug text,
    schema_name text,
    tenant_set_sha256 bytea
)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path TO pg_catalog
AS $fn$
DECLARE matched bigint;
BEGIN
    IF p_purpose IS DISTINCT FROM 'backup' AND p_purpose IS DISTINCT FROM 'restore' THEN
        RAISE EXCEPTION 'tenancy recovery target purpose is invalid' USING ERRCODE = '22023';
    END IF;
    SELECT count(*) INTO matched
    FROM platform.tenants AS tenant
    JOIN platform.tenancy_activation_facts AS fact
        ON fact.tenant_id = tenant.id
        AND fact.schema_name = tenant.schema_name
        AND fact.slug = tenant.slug
    JOIN platform.tenancy_onboarding_operations AS operation
        ON operation.operation_id = fact.operation_id
        AND operation.tenant_id = tenant.id
        AND operation.environment_id = fact.environment_id
        AND operation.schema_name = tenant.schema_name
    WHERE tenant.deleted_at IS NULL
        AND tenant.purged_at IS NULL
        AND tenant.status = 'ACTIVE'
        AND operation.state = 'succeeded';
    IF matched <> 1 THEN
        RAISE EXCEPTION 'tenancy recovery target is missing' USING ERRCODE = '02000';
    END IF;
    RETURN QUERY
    SELECT fact.environment_id, tenant.id, tenant.slug, tenant.schema_name,
        pg_catalog.sha256(pg_catalog.convert_to(
            'tenant-set.v1|' || tenant.id::text || '|' || tenant.schema_name, 'UTF8'))
    FROM platform.tenants AS tenant
    JOIN platform.tenancy_activation_facts AS fact
        ON fact.tenant_id = tenant.id
        AND fact.schema_name = tenant.schema_name
        AND fact.slug = tenant.slug
    JOIN platform.tenancy_onboarding_operations AS operation
        ON operation.operation_id = fact.operation_id
        AND operation.tenant_id = tenant.id
        AND operation.environment_id = fact.environment_id
        AND operation.schema_name = tenant.schema_name
    WHERE tenant.deleted_at IS NULL
        AND tenant.purged_at IS NULL
        AND tenant.status = 'ACTIVE'
        AND operation.state = 'succeeded';
END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_get_recovery_target_v1(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_get_recovery_target_v1(text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_get_recovery_target_v1(text);
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_get_recovery_target_v1(p_purpose text)
RETURNS TABLE (environment_id uuid, tenant_id uuid, tenant_slug text, schema_name text, tenant_set_sha256 bytea)
LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$
    SELECT '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text, ''::text,
        decode('0000000000000000000000000000000000000000000000000000000000000000', 'hex');
$fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.tenancy_get_recovery_target_v1(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_get_recovery_target_v1(text) TO harpy_api, harpy_worker;
