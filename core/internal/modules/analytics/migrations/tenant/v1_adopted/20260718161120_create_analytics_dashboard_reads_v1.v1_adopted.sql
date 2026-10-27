-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_widgets_v1(p_version_id uuid)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    RETURN (SELECT COALESCE(pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
        'widget_id', widget.widget_id,
        'position', widget.position,
        'query_version_id', widget.query_version_id,
        'parameter_document', pg_catalog.encode(widget.parameter_document, 'hex'),
        'output_document', pg_catalog.encode(widget.output_document, 'hex')
    ) ORDER BY widget.position), '[]'::jsonb)
    FROM analytics_dashboard_widgets AS widget
    WHERE widget.dashboard_version_id = p_version_id);
END;
$function$;

CREATE FUNCTION analytics_dashboard_get_version_v1(
    p_scope_schema name, p_dashboard_id uuid, p_version_id uuid
)
RETURNS TABLE (
    dashboard_id uuid, version_id uuid, version_no bigint,
    root_revision bigint, version_revision bigint, lifecycle_state text,
    version_state text, name text, visibility text, layout_document bytea,
    layout_sha256 bytea, manifest_sha256 bytea, widgets jsonb,
    created_at timestamptz, updated_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    RETURN QUERY
    SELECT root.id, version_row.id, version_row.version_no, root.row_revision,
           version_row.row_revision, root.lifecycle_state, version_row.version_state,
           version_row.name, version_row.visibility, version_row.layout_document,
           version_row.layout_sha256, version_row.manifest_sha256,
           analytics_dashboard_widgets_v1(version_row.id),
           root.created_at, version_row.updated_at
    FROM analytics_dashboards AS root
    JOIN analytics_dashboard_versions AS version_row
      ON version_row.dashboard_id = root.id
    WHERE root.id = p_dashboard_id AND version_row.id = p_version_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102',
            MESSAGE = 'analytics dashboard version was not found';
    END IF;
END;
$function$;

CREATE FUNCTION analytics_dashboard_get_v1(p_scope_schema name, p_dashboard_id uuid)
RETURNS TABLE (
    dashboard_id uuid, version_id uuid, version_no bigint,
    root_revision bigint, version_revision bigint, lifecycle_state text,
    version_state text, name text, visibility text, layout_document bytea,
    layout_sha256 bytea, manifest_sha256 bytea, widgets jsonb,
    created_at timestamptz, updated_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_version_id uuid;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT current_draft_version_id INTO v_version_id
    FROM analytics_dashboards WHERE id = p_dashboard_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102', MESSAGE = 'analytics dashboard was not found';
    END IF;
    RETURN QUERY SELECT * FROM analytics_dashboard_get_version_v1(
        p_scope_schema, p_dashboard_id, v_version_id);
END;
$function$;

CREATE FUNCTION analytics_dashboard_list_v1(p_scope_schema name, p_limit integer, p_after uuid)
RETURNS TABLE (
    dashboard_id uuid, version_id uuid, version_no bigint,
    root_revision bigint, version_revision bigint, lifecycle_state text,
    version_state text, name text, visibility text, layout_document bytea,
    layout_sha256 bytea, manifest_sha256 bytea, widgets jsonb,
    created_at timestamptz, updated_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_limit IS NULL OR p_limit NOT BETWEEN 1 AND 100 THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'analytics dashboard list input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    RETURN QUERY
    SELECT root.id, version_row.id, version_row.version_no, root.row_revision,
           version_row.row_revision, root.lifecycle_state, version_row.version_state,
           version_row.name, version_row.visibility, version_row.layout_document,
           version_row.layout_sha256, version_row.manifest_sha256,
           analytics_dashboard_widgets_v1(version_row.id), root.created_at, root.updated_at
    FROM analytics_dashboards AS root
    JOIN analytics_dashboard_versions AS version_row
      ON version_row.id = root.current_draft_version_id AND version_row.dashboard_id = root.id
    WHERE root.lifecycle_state = 'active' AND (p_after IS NULL OR root.id > p_after)
    ORDER BY root.id LIMIT p_limit;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_widgets_v1(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_dashboard_get_version_v1(name,uuid,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_dashboard_get_v1(name,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_dashboard_list_v1(name,integer,uuid) FROM PUBLIC;
