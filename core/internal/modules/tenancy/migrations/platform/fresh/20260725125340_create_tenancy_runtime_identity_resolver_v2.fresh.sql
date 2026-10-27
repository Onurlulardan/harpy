-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_runtime_resolve_identity_v2(p_tenant_id uuid)
RETURNS TABLE (
    schema_name name,
    slug text,
    status text,
    not_deleted boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_tenant_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Tenant runtime identity resolver denied';
    END IF;
    RETURN QUERY SELECT
        tenant.schema_name::name,
        tenant.slug,
        tenant.status::text,
        tenant.deleted_at IS NULL
    FROM platform.tenants AS tenant
    WHERE tenant.id = p_tenant_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.tenancy_runtime_resolve_identity_v2(uuid)
FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
    platform.tenancy_runtime_resolve_identity_v2(uuid)
TO forgecrud_runtime;
