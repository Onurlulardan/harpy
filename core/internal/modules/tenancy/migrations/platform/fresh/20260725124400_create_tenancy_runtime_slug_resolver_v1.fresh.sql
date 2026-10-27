-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_runtime_resolve_slug_v1(p_slug text)
RETURNS TABLE (
    id uuid,
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
       p_slug IS NULL OR p_slug = '' OR pg_catalog.octet_length(p_slug) > 400 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Tenant runtime slug resolver denied';
    END IF;
    RETURN QUERY SELECT
        tenant.id,
        tenant.status::text,
        tenant.deleted_at IS NULL
    FROM platform.tenants AS tenant
    WHERE tenant.slug COLLATE "C" = p_slug COLLATE "C";
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.tenancy_runtime_resolve_slug_v1(text)
FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
    platform.tenancy_runtime_resolve_slug_v1(text)
TO forgecrud_runtime;
