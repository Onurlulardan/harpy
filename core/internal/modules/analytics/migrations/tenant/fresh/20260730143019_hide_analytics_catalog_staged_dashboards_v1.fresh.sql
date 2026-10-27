-- +goose Up
DROP FUNCTION analytics_dashboard_get_version_v1(name,uuid,uuid);
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_get_version_v1(
    p_scope_schema name,p_dashboard_id uuid,p_version_id uuid
)
RETURNS TABLE(
    dashboard_id uuid,version_id uuid,version_no bigint,
    root_revision bigint,version_revision bigint,lifecycle_state text,
    version_state text,name text,visibility text,layout_document bytea,
    layout_sha256 bytea,manifest_sha256 bytea,widgets jsonb,
    created_at timestamptz,updated_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    RETURN QUERY
    SELECT root.id,version_row.id,version_row.version_no,root.row_revision,
        version_row.row_revision,root.lifecycle_state,
        version_row.version_state,version_row.name,version_row.visibility,
        version_row.layout_document,version_row.layout_sha256,
        version_row.manifest_sha256,
        analytics_dashboard_widgets_v1(version_row.id),
        root.created_at,version_row.updated_at
    FROM analytics_dashboards AS root
    JOIN analytics_dashboard_versions AS version_row
      ON version_row.dashboard_id=root.id
    WHERE root.id=p_dashboard_id AND version_row.id=p_version_id
      AND root.lifecycle_state<>'staged';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='analytics dashboard version was not found';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_dashboard_get_version_v1(
    name,uuid,uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_get_version_v1(
    name,uuid,uuid
) TO forgecrud_runtime;
