-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_retention_purge_versions_v1(
    p_cutoff timestamptz,p_limit integer
)
RETURNS TABLE (query_versions_purged bigint,dashboard_versions_purged bigint)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_query_versions jsonb;
    v_dashboard_versions jsonb;
    v_remaining integer;
BEGIN
    IF p_cutoff IS NULL OR NOT pg_catalog.isfinite(p_cutoff) OR
       p_limit NOT BETWEEN 1 AND 100 OR
       NOT analytics_retention_delete_allowed_v1(pg_catalog.pg_backend_pid()) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics version retention authority is invalid';
    END IF;
    SELECT COALESCE(pg_catalog.jsonb_agg(candidate.version_id::text),'[]'::jsonb)
    INTO v_query_versions
    FROM (
        SELECT version_row.id AS version_id
        FROM analytics_saved_query_versions AS version_row
        WHERE version_row.version_state = 'retired' AND version_row.updated_at <= p_cutoff
          AND NOT EXISTS (
              SELECT 1 FROM analytics_inbound_references AS reference_row
              WHERE reference_row.target_kind = 'query_version'
                AND reference_row.target_id = version_row.id
                AND reference_row.state = 'active'
          )
          AND NOT EXISTS (
              SELECT 1 FROM analytics_publication_releases AS release
              WHERE release.resource_kind = 'query'
                AND release.resource_version_id = version_row.id
          )
          AND NOT EXISTS (
              SELECT 1 FROM analytics_dashboard_widgets AS widget
              WHERE widget.query_version_id = version_row.id
          )
          AND NOT EXISTS (
              SELECT 1 FROM analytics_saved_queries AS root
              WHERE root.current_draft_version_id = version_row.id
          )
        ORDER BY version_row.updated_at,version_row.id
        FOR UPDATE OF version_row SKIP LOCKED LIMIT p_limit
    ) AS candidate;
    DELETE FROM analytics_query_dependencies
    WHERE v_query_versions ? query_version_id::text;
    DELETE FROM analytics_saved_query_versions
    WHERE v_query_versions ? id::text;
    GET DIAGNOSTICS query_versions_purged = ROW_COUNT;
    DELETE FROM analytics_saved_queries AS root
    WHERE root.lifecycle_state = 'retired' AND root.current_draft_version_id IS NULL
      AND NOT EXISTS (
          SELECT 1 FROM analytics_saved_query_versions AS version_row
          WHERE version_row.query_id = root.id
      );
    v_remaining := p_limit - query_versions_purged::integer;
    IF v_remaining < 1 THEN
        dashboard_versions_purged := 0;
        RETURN NEXT;
        RETURN;
    END IF;
    SELECT COALESCE(pg_catalog.jsonb_agg(candidate.version_id::text),'[]'::jsonb)
    INTO v_dashboard_versions
    FROM (
        SELECT version_row.id AS version_id
        FROM analytics_dashboard_versions AS version_row
        WHERE version_row.version_state = 'retired' AND version_row.updated_at <= p_cutoff
          AND NOT EXISTS (
              SELECT 1 FROM analytics_inbound_references AS reference_row
              WHERE reference_row.target_kind = 'dashboard_version'
                AND reference_row.target_id = version_row.id
                AND reference_row.state = 'active'
          )
          AND NOT EXISTS (
              SELECT 1 FROM analytics_publication_releases AS release
              WHERE release.resource_kind = 'dashboard'
                AND release.resource_version_id = version_row.id
          )
          AND NOT EXISTS (
              SELECT 1 FROM analytics_dashboards AS root
              WHERE root.current_draft_version_id = version_row.id
          )
        ORDER BY version_row.updated_at,version_row.id
        FOR UPDATE OF version_row SKIP LOCKED LIMIT v_remaining
    ) AS candidate;
    DELETE FROM analytics_dashboard_widgets
    WHERE v_dashboard_versions ? dashboard_version_id::text;
    DELETE FROM analytics_dashboard_versions
    WHERE v_dashboard_versions ? id::text;
    GET DIAGNOSTICS dashboard_versions_purged = ROW_COUNT;
    DELETE FROM analytics_dashboards AS root
    WHERE root.lifecycle_state = 'retired' AND root.current_draft_version_id IS NULL
      AND NOT EXISTS (
          SELECT 1 FROM analytics_dashboard_versions AS version_row
          WHERE version_row.dashboard_id = root.id
      );
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_retention_purge_versions_v1(
    timestamptz,integer
) FROM PUBLIC;
