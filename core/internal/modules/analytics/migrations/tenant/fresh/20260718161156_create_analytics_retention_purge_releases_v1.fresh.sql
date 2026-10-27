-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_retention_purge_releases_v1(
    p_cutoff timestamptz,p_limit integer
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_release_ids jsonb;
    v_eligible_ids jsonb;
    v_release_text text;
    v_tenant_id uuid;
    v_purged bigint;
BEGIN
    IF p_cutoff IS NULL OR NOT pg_catalog.isfinite(p_cutoff) OR
       p_limit NOT BETWEEN 1 AND 100 OR
       NOT analytics_retention_delete_allowed_v1(pg_catalog.pg_backend_pid()) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics release retention authority is invalid';
    END IF;
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity_row
    WHERE identity_row.singleton
      AND identity_row.schema_name = pg_catalog.current_schema();
    SELECT COALESCE(pg_catalog.jsonb_agg(candidate.release_id::text),'[]'::jsonb)
    INTO v_release_ids
    FROM (
        SELECT release.id AS release_id
        FROM analytics_publication_releases AS release
        WHERE release.lifecycle_state IN ('revoked','retired')
          AND release.revoked_at <= p_cutoff
          AND NOT EXISTS (
              SELECT 1 FROM analytics_inbound_references AS reference_row
              WHERE reference_row.target_kind = 'publication_release'
                AND reference_row.target_id = release.id
                AND reference_row.state = 'active'
          )
          AND NOT EXISTS (
              SELECT 1 FROM analytics_execution_audits AS execution
              WHERE execution.release_id = release.id
          )
          AND NOT EXISTS (
              SELECT 1 FROM analytics_capability_tokens AS token
              WHERE token.release_id = release.id
          )
        ORDER BY release.revoked_at,release.id
        FOR UPDATE OF release SKIP LOCKED LIMIT p_limit
    ) AS candidate;
    v_eligible_ids := '[]'::jsonb;
    FOR v_release_text IN
        SELECT pg_catalog.jsonb_array_elements_text(v_release_ids)
    LOOP
        IF platform.analytics_public_retention_release_v1(
            v_tenant_id,v_release_text::uuid,p_cutoff
        ) THEN
            v_eligible_ids := v_eligible_ids || pg_catalog.jsonb_build_array(v_release_text);
        END IF;
    END LOOP;
    DELETE FROM analytics_publication_releases
    WHERE v_eligible_ids ? id::text;
    GET DIAGNOSTICS v_purged = ROW_COUNT;
    RETURN v_purged;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_retention_purge_releases_v1(
    timestamptz,integer
) FROM PUBLIC;
