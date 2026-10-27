-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.analytics_legacy_public_record_hit_v1(
    p_tenant_identity text, p_method text, p_path text, p_slug text,
    p_tenant_id uuid, p_legacy_endpoint_id uuid, p_public_id bytea,
    p_release_id uuid, p_hit_at timestamptz
)
RETURNS TABLE (
    legacy_tenant_identity text, method text, path text, slug text,
    tenant_id uuid, legacy_endpoint_id uuid, public_id bytea,
    release_id uuid, state text, hit_count bigint, last_hit_at timestamptz,
    consumers_acked_at timestamptz, remove_after timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_hit_at IS NULL OR NOT pg_catalog.isfinite(p_hit_at) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics legacy hit authority is invalid';
    END IF;
    RETURN QUERY
    UPDATE platform.analytics_legacy_public_routes AS route
       SET hit_count = route.hit_count + 1, last_hit_at = p_hit_at
    WHERE route.legacy_tenant_identity = p_tenant_identity
      AND route.method = p_method AND route.path = p_path AND route.slug = p_slug
      AND route.tenant_id = p_tenant_id
      AND route.legacy_endpoint_id = p_legacy_endpoint_id
      AND route.public_id IS NOT DISTINCT FROM p_public_id
      AND route.release_id = p_release_id AND route.state = 'active'
      AND EXISTS (
          SELECT 1 FROM platform.analytics_publications AS publication
          WHERE publication.public_id = route.public_id
            AND publication.tenant_id = route.tenant_id
            AND publication.release_id = route.release_id
            AND publication.state = 'active'
      )
    RETURNING route.legacy_tenant_identity, route.method, route.path, route.slug,
              route.tenant_id, route.legacy_endpoint_id, route.public_id,
              route.release_id, route.state, route.hit_count, route.last_hit_at,
              route.consumers_acked_at, route.remove_after;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102',
            MESSAGE = 'analytics legacy hit mapping was not found';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.analytics_legacy_public_record_hit_v1(
    text,text,text,text,uuid,uuid,bytea,uuid,timestamptz
) FROM PUBLIC;
