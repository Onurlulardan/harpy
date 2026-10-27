-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_public_query_root_v1(
    p_scope_schema name, p_release_id uuid, p_query_version_id uuid
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE v_query_id uuid;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT version_row.query_id INTO v_query_id
    FROM analytics_publication_releases AS release
    JOIN analytics_saved_query_versions AS version_row
      ON version_row.id = p_query_version_id
     AND version_row.version_state = 'published'
    WHERE release.id = p_release_id
      AND release.lifecycle_state = 'active'
      AND release.audience IN ('public', 'capability')
      AND (
          (release.resource_kind = 'query'
           AND release.resource_version_id = p_query_version_id)
          OR
          (release.resource_kind = 'dashboard' AND EXISTS (
              SELECT 1 FROM analytics_dashboard_widgets AS widget
              WHERE widget.dashboard_version_id = release.resource_version_id
                AND widget.query_version_id = p_query_version_id
          ))
      );
    IF v_query_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102',
            MESSAGE = 'analytics public query root was not found';
    END IF;
    RETURN v_query_id;
END;
$function$;

CREATE FUNCTION analytics_public_dashboard_root_v1(
    p_scope_schema name, p_release_id uuid, p_dashboard_version_id uuid
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE v_dashboard_id uuid;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT version_row.dashboard_id INTO v_dashboard_id
    FROM analytics_publication_releases AS release
    JOIN analytics_dashboard_versions AS version_row
      ON version_row.id = p_dashboard_version_id
     AND version_row.version_state = 'published'
    WHERE release.id = p_release_id
      AND release.lifecycle_state = 'active'
      AND release.audience IN ('public', 'capability')
      AND release.resource_kind = 'dashboard'
      AND release.resource_version_id = p_dashboard_version_id;
    IF v_dashboard_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102',
            MESSAGE = 'analytics public dashboard root was not found';
    END IF;
    RETURN v_dashboard_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_public_query_root_v1(name,uuid,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_public_dashboard_root_v1(name,uuid,uuid) FROM PUBLIC;
