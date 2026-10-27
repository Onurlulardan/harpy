-- +goose Up
ALTER TABLE platform.analytics_legacy_public_routes
    ADD COLUMN consumer_inventory_sha256 bytea,
    ADD COLUMN retirement_evidence_sha256 bytea,
    ADD COLUMN retirement_observed_at timestamptz,
    ADD COLUMN retired_at timestamptz,
    ADD CONSTRAINT analytics_legacy_public_routes_consumer_check CHECK (
        (consumers_acked_at IS NULL AND consumer_inventory_sha256 IS NULL)
        OR (consumers_acked_at IS NOT NULL
            AND pg_catalog.octet_length(consumer_inventory_sha256) = 32)
    ),
    ADD CONSTRAINT analytics_legacy_public_routes_retirement_check CHECK (
        (state = 'active' AND retirement_evidence_sha256 IS NULL
            AND retirement_observed_at IS NULL AND retired_at IS NULL)
        OR (state = 'retired'
            AND pg_catalog.octet_length(retirement_evidence_sha256) = 32
            AND pg_catalog.isfinite(retirement_observed_at)
            AND pg_catalog.isfinite(retired_at)
            AND retirement_observed_at <= retired_at)
    );

-- +goose StatementBegin
CREATE FUNCTION platform.analytics_legacy_public_resolve_v1(
    p_tenant_identity text, p_method text, p_path text, p_slug text
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
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics legacy resolver authority is invalid';
    END IF;
    RETURN QUERY
    SELECT route.legacy_tenant_identity, route.method, route.path, route.slug,
           route.tenant_id, route.legacy_endpoint_id, route.public_id,
           route.release_id, route.state, route.hit_count, route.last_hit_at,
           route.consumers_acked_at, route.remove_after
    FROM platform.analytics_legacy_public_routes AS route
    JOIN platform.analytics_publications AS publication
      ON publication.public_id = route.public_id
     AND publication.tenant_id = route.tenant_id
     AND publication.release_id = route.release_id
    WHERE route.legacy_tenant_identity = p_tenant_identity
      AND route.method = p_method AND route.path = p_path AND route.slug = p_slug
      AND route.state = 'active' AND publication.state = 'active';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102',
            MESSAGE = 'analytics legacy mapping was not found';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.analytics_legacy_public_resolve_v1(
    text,text,text,text
) FROM PUBLIC;
