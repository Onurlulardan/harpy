-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.analytics_public_repair_v1(
    p_tenant_id uuid, p_public_id bytea, p_release_id uuid,
    p_resource_kind text, p_resource_version_id uuid, p_audience text,
    p_state text, p_generation bigint, p_capability_sha256 bytea,
    p_capability_expires_at timestamptz, p_activated_at timestamptz,
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
       pg_catalog.octet_length(p_public_id) IS DISTINCT FROM 16 OR
       pg_catalog.uuid_extract_version(p_release_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_resource_version_id) IS DISTINCT FROM 7 OR
       p_resource_kind NOT IN ('query','dashboard') OR
       p_audience NOT IN ('public','capability') OR
       p_state NOT IN ('active','revoked') OR p_generation < 1 OR
       p_activated_at IS NULL OR NOT pg_catalog.isfinite(p_activated_at) OR
       (p_state = 'active' AND p_revoked_at IS NOT NULL) OR
       (p_state = 'revoked' AND
        (p_revoked_at IS NULL OR NOT pg_catalog.isfinite(p_revoked_at) OR
         p_revoked_at < p_activated_at)) OR
       (p_audience = 'public' AND
        (p_capability_sha256 IS NOT NULL OR p_capability_expires_at IS NOT NULL)) OR
       (p_audience = 'capability' AND
        (pg_catalog.octet_length(p_capability_sha256) IS DISTINCT FROM 32 OR
         p_capability_expires_at IS NULL OR
         NOT pg_catalog.isfinite(p_capability_expires_at) OR
         p_capability_expires_at <= p_activated_at)) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics public repair input is invalid';
    END IF;
    PERFORM 1
    FROM platform.analytics_resolve_current_tenant_authority_v1(p_tenant_id);
    INSERT INTO platform.analytics_publications (
        public_id,tenant_id,release_id,resource_kind,resource_version_id,
        audience,state,generation,capability_sha256,capability_expires_at,
        activated_at,revoked_at
    ) VALUES (
        p_public_id,p_tenant_id,p_release_id,p_resource_kind,
        p_resource_version_id,p_audience,p_state,p_generation,
        p_capability_sha256,p_capability_expires_at,p_activated_at,p_revoked_at
    )
    ON CONFLICT (public_id) DO UPDATE
    SET state=EXCLUDED.state,generation=EXCLUDED.generation,
        capability_sha256=EXCLUDED.capability_sha256,
        capability_expires_at=EXCLUDED.capability_expires_at,
        activated_at=EXCLUDED.activated_at,revoked_at=EXCLUDED.revoked_at
    WHERE platform.analytics_publications.tenant_id=EXCLUDED.tenant_id
      AND platform.analytics_publications.release_id=EXCLUDED.release_id
      AND platform.analytics_publications.resource_kind=EXCLUDED.resource_kind
      AND platform.analytics_publications.resource_version_id=EXCLUDED.resource_version_id
      AND platform.analytics_publications.audience=EXCLUDED.audience
      AND platform.analytics_publications.generation <= EXCLUDED.generation;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='analytics public repair identity conflicts';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.analytics_public_repair_v1(
    uuid,bytea,uuid,text,uuid,text,text,bigint,bytea,timestamptz,
    timestamptz,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.analytics_public_repair_v1(
    uuid,bytea,uuid,text,uuid,text,text,bigint,bytea,timestamptz,
    timestamptz,timestamptz
) TO harpy_runtime;

