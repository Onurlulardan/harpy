-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.analytics_public_resolve_v1(p_public_id bytea)
RETURNS TABLE (
    public_id bytea, tenant_id uuid, release_id uuid, resource_kind text,
    resource_version_id uuid, audience text, state text, generation bigint,
    capability_sha256 bytea, capability_expires_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF pg_catalog.octet_length(p_public_id) IS DISTINCT FROM 16 THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'analytics public identity is invalid';
    END IF;
    RETURN QUERY SELECT publication.public_id, publication.tenant_id,
        publication.release_id, publication.resource_kind,
        publication.resource_version_id, publication.audience,
        publication.state, publication.generation, publication.capability_sha256,
        publication.capability_expires_at
    FROM platform.analytics_publications AS publication
    WHERE publication.public_id = p_public_id AND publication.state = 'active';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102', MESSAGE = 'analytics public route was not found';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.analytics_public_resolve_v1(bytea) FROM PUBLIC;
