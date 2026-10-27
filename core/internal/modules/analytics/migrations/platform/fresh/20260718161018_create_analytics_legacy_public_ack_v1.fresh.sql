-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.analytics_legacy_public_ack_consumers_v1(
    p_tenant_identity text, p_method text, p_path text, p_slug text,
    p_tenant_id uuid, p_legacy_endpoint_id uuid, p_public_id bytea,
    p_release_id uuid, p_inventory_sha256 bytea, p_acked_at timestamptz
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
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_cluster_admin' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       pg_catalog.octet_length(p_inventory_sha256) IS DISTINCT FROM 32 OR
       p_acked_at IS NULL OR NOT pg_catalog.isfinite(p_acked_at) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics consumer acknowledgement authority is invalid';
    END IF;
    RETURN QUERY
    UPDATE platform.analytics_legacy_public_routes AS route
       SET consumers_acked_at = p_acked_at,
           consumer_inventory_sha256 = p_inventory_sha256
    WHERE route.legacy_tenant_identity = p_tenant_identity
      AND route.method = p_method AND route.path = p_path AND route.slug = p_slug
      AND route.tenant_id = p_tenant_id
      AND route.legacy_endpoint_id = p_legacy_endpoint_id
      AND route.public_id IS NOT DISTINCT FROM p_public_id
      AND route.release_id = p_release_id AND route.state = 'active'
      AND (route.consumers_acked_at IS NULL OR (
          route.consumers_acked_at IS NOT DISTINCT FROM p_acked_at
          AND route.consumer_inventory_sha256 IS NOT DISTINCT FROM p_inventory_sha256
      ))
    RETURNING route.legacy_tenant_identity, route.method, route.path, route.slug,
              route.tenant_id, route.legacy_endpoint_id, route.public_id,
              route.release_id, route.state, route.hit_count, route.last_hit_at,
              route.consumers_acked_at, route.remove_after;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6103',
            MESSAGE = 'analytics consumer acknowledgement conflicts';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.analytics_legacy_public_ack_consumers_v1(
    text,text,text,text,uuid,uuid,bytea,uuid,bytea,timestamptz
) FROM PUBLIC;
