-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.analytics_public_revoke_v1(
    p_tenant_id uuid, p_release_id uuid, p_generation bigint,
    p_revoked_at timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       p_tenant_id IS NULL OR
       p_tenant_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.uuid_extract_version(p_release_id) IS DISTINCT FROM 7 OR
       p_generation < 1 OR p_revoked_at IS NULL OR
       NOT pg_catalog.isfinite(p_revoked_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics public revoke input is invalid';
    END IF;
    PERFORM 1
    FROM platform.analytics_resolve_current_tenant_authority_v1(p_tenant_id);
    UPDATE platform.analytics_publications
    SET state='revoked',generation=generation+1,revoked_at=p_revoked_at
    WHERE tenant_id=p_tenant_id AND release_id=p_release_id
      AND generation=p_generation AND state='active';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='analytics public pointer changed';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.analytics_public_revoke_v1(
    uuid,uuid,bigint,timestamptz
) FROM PUBLIC;
