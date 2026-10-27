-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.analytics_legacy_public_retire_v1(
    p_tenant_identity text, p_method text, p_path text, p_slug text,
    p_tenant_id uuid, p_legacy_endpoint_id uuid, p_public_id bytea,
    p_release_id uuid, p_evidence_sha256 bytea,
    p_observed_at timestamptz, p_retired_at timestamptz
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
       pg_catalog.octet_length(p_evidence_sha256) IS DISTINCT FROM 32 OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) OR
       p_retired_at IS NULL OR NOT pg_catalog.isfinite(p_retired_at) OR
       p_observed_at > p_retired_at OR
       p_retired_at - p_observed_at > INTERVAL '5 minutes' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics legacy retirement authority is invalid';
    END IF;
    IF EXISTS (
        SELECT 1
        FROM platform.tenants AS tenant
        LEFT JOIN platform.domain_cutovers AS cutover
          ON cutover.scope_kind = 'tenant' AND cutover.scope_id = tenant.id
         AND cutover.domain = 'analytics'
        WHERE tenant.deleted_at IS NULL
          AND (cutover.scope_id IS NULL OR cutover.state <> 'v2' OR cutover.writer <> 'v2')
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P6104',
            MESSAGE = 'analytics legacy retirement awaits tenant cutover';
    END IF;
    RETURN QUERY
    UPDATE platform.analytics_legacy_public_routes AS route
       SET state = 'retired', retirement_evidence_sha256 = p_evidence_sha256,
           retirement_observed_at = p_observed_at, retired_at = p_retired_at
    WHERE route.legacy_tenant_identity = p_tenant_identity
      AND route.method = p_method AND route.path = p_path AND route.slug = p_slug
      AND route.tenant_id = p_tenant_id
      AND route.legacy_endpoint_id = p_legacy_endpoint_id
      AND route.public_id IS NOT DISTINCT FROM p_public_id
      AND route.release_id = p_release_id AND route.state = 'active'
      AND route.consumers_acked_at IS NOT NULL
      AND route.consumer_inventory_sha256 IS NOT NULL
      AND route.remove_after <= p_retired_at
      AND (route.last_hit_at IS NULL
           OR route.last_hit_at <= p_retired_at - INTERVAL '30 days')
    RETURNING route.legacy_tenant_identity, route.method, route.path, route.slug,
              route.tenant_id, route.legacy_endpoint_id, route.public_id,
              route.release_id, route.state, route.hit_count, route.last_hit_at,
              route.consumers_acked_at, route.remove_after;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6104',
            MESSAGE = 'analytics legacy retirement gate is closed';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.analytics_legacy_public_retire_v1(
    text,text,text,text,uuid,uuid,bytea,uuid,bytea,timestamptz,timestamptz
) FROM PUBLIC;
