-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.analytics_public_retention_release_v1(
    p_tenant_id uuid,p_release_id uuid,p_cutoff timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       p_tenant_id IS NULL OR
       p_tenant_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.uuid_extract_version(p_release_id) IS DISTINCT FROM 7 OR
       p_cutoff IS NULL OR NOT pg_catalog.isfinite(p_cutoff) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'analytics public retention input is invalid';
    END IF;
    PERFORM 1 FROM platform.analytics_resolve_current_tenant_authority_v1(p_tenant_id);
    IF EXISTS (
        SELECT 1 FROM platform.analytics_legacy_public_routes AS legacy
        WHERE legacy.tenant_id = p_tenant_id AND legacy.release_id = p_release_id
    ) OR EXISTS (
        SELECT 1 FROM platform.analytics_publications AS pointer
        WHERE pointer.tenant_id = p_tenant_id AND pointer.release_id = p_release_id
          AND (pointer.state <> 'revoked' OR pointer.revoked_at > p_cutoff)
    ) THEN
        RETURN false;
    END IF;
    DELETE FROM platform.analytics_public_rate_windows
    WHERE tenant_id = p_tenant_id AND release_id = p_release_id;
    DELETE FROM platform.analytics_public_concurrency_leases
    WHERE tenant_id = p_tenant_id AND release_id = p_release_id;
    DELETE FROM platform.analytics_publications
    WHERE tenant_id = p_tenant_id AND release_id = p_release_id
      AND state = 'revoked' AND revoked_at <= p_cutoff;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.analytics_public_retention_release_v1(
    uuid,uuid,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.analytics_public_retention_release_v1(
    uuid,uuid,timestamptz
) TO harpy_runtime;
