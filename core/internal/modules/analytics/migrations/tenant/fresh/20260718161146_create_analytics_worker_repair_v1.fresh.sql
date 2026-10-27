-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_publication_repair_v1(
    p_scope_schema name, p_release_id uuid, p_expected_generation bigint,
    p_observed_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_release analytics_publication_releases%ROWTYPE;
    v_tenant_id uuid;
    v_capability_sha256 bytea;
BEGIN
    IF pg_catalog.uuid_extract_version(p_release_id) IS DISTINCT FROM 7 OR
       p_expected_generation < 1 OR p_observed_at IS NULL OR
       NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics publication repair input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT release.* INTO v_release
    FROM analytics_publication_releases AS release
    WHERE release.id=p_release_id
      AND release.generation=p_expected_generation
      AND release.audience IN ('public','capability')
      AND release.lifecycle_state IN ('active','revoked')
      AND release.public_id IS NOT NULL
      AND release.activated_at IS NOT NULL
      AND release.updated_at <= p_observed_at
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='analytics publication repair target changed';
    END IF;
    SELECT tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity
    WHERE singleton AND schema_name=pg_catalog.current_schema();
    IF v_release.audience='capability' THEN
        SELECT token.token_sha256 INTO STRICT v_capability_sha256
        FROM analytics_capability_tokens AS token
        WHERE token.release_id=v_release.id;
    END IF;
    PERFORM platform.analytics_public_repair_v1(
        v_tenant_id,v_release.public_id,v_release.id,v_release.resource_kind,
        v_release.resource_version_id,v_release.audience,
        v_release.lifecycle_state,v_release.generation,v_capability_sha256,
        v_release.capability_expires_at,v_release.activated_at,
        v_release.revoked_at
    );
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_publication_repair_v1(
    name,uuid,bigint,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_publication_repair_v1(
    name,uuid,bigint,timestamptz
) TO harpy_runtime;

