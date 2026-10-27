-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION platform.tenancy_get_recovery_target_v1(p_purpose text)
RETURNS TABLE(
    environment_id uuid,
    deployment_mode text,
    tenant_id uuid,
    tenant_slug text,
    schema_name name,
    tenant_set_sha256 bytea
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF NOT (
        SESSION_USER = 'forgecrud_backup' AND p_purpose = 'backup'
        OR SESSION_USER = 'forgecrud_restore' AND p_purpose = 'restore'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'tenancy recovery target denied';
    END IF;
    RETURN QUERY
    WITH tenant_set AS (
        SELECT policy.environment_id,
               pg_catalog.sha256(pg_catalog.convert_to(
                   'tenant-set.v1' || COALESCE(pg_catalog.string_agg(
                       '|' || binding.tenant_id::text || '|' ||
                       tenant.schema_name,
                       '' ORDER BY binding.tenant_id::text
                   ), ''),
                   'UTF8'
               )) AS digest
        FROM platform.tenancy_deployment_policy AS policy
        JOIN platform.tenancy_tenant_bindings AS binding
          ON binding.environment_id = policy.environment_id
        JOIN platform.tenants AS tenant ON tenant.id = binding.tenant_id
        WHERE policy.singleton AND policy.deployment_mode = 'onpremise'
          AND binding.purpose = 'onpremise_primary'
          AND tenant.status = 'ACTIVE' AND tenant.deleted_at IS NULL
        GROUP BY policy.environment_id
    )
    SELECT policy.environment_id, policy.deployment_mode,
           binding.tenant_id, tenant.slug::text, tenant.schema_name::name,
           tenant_set.digest
    FROM platform.tenancy_deployment_policy AS policy
    JOIN platform.tenancy_tenant_bindings AS binding
      ON binding.environment_id = policy.environment_id
     AND binding.purpose = 'onpremise_primary'
    JOIN platform.tenants AS tenant ON tenant.id = binding.tenant_id
    JOIN tenant_set ON tenant_set.environment_id = policy.environment_id
    WHERE policy.singleton AND policy.deployment_mode = 'onpremise'
      AND tenant.status = 'ACTIVE' AND tenant.deleted_at IS NULL;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_get_recovery_target_v1(text) FROM PUBLIC;
GRANT USAGE ON SCHEMA platform TO forgecrud_backup_reader, forgecrud_restore;
GRANT EXECUTE ON FUNCTION platform.tenancy_get_recovery_target_v1(text)
TO forgecrud_backup_reader, forgecrud_restore;

