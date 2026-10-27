-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_forms_import_source_kek_tenant_v1(
    p_after_tenant_id uuid
)
RETURNS TABLE(tenant_id uuid, tenant_schema name)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Forms import-source KEK inventory authority is invalid';
    END IF;
    RETURN QUERY
    SELECT tenant.id,tenant.schema_name::name
    FROM platform.tenants AS tenant
    JOIN platform.domain_cutovers AS authority
      ON authority.scope_kind='tenant' AND authority.scope_id=tenant.id
     AND authority.domain='forms' AND authority.state='v2'
     AND authority.writer='v2'
    WHERE tenant.deleted_at IS NULL AND tenant.status IN ('ACTIVE','TRIAL')
      AND (p_after_tenant_id IS NULL OR tenant.id > p_after_tenant_id)
    ORDER BY tenant.id
    LIMIT 1;
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.platform_async_forms_import_source_kek_tenant_v1(uuid) FROM PUBLIC;
-- +goose StatementEnd
