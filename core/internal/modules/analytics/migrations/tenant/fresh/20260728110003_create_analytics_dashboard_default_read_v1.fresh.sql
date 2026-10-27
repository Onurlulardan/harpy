-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_get_default_v1(
    p_scope_schema name, p_principal_id uuid, p_surface text
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
DECLARE
    v_dashboard_id uuid;
BEGIN
    IF p_principal_id IS NULL OR
       p_principal_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       p_surface IS NULL OR
       p_surface OPERATOR(pg_catalog.!~) '^[a-z][a-z0-9_.-]{0,63}$' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'analytics dashboard default read input is invalid';
    END IF;

    PERFORM analytics_assert_current_authority_v1(p_scope_schema);

    SELECT preference.dashboard_id
    INTO v_dashboard_id
    FROM analytics_dashboard_preferences AS preference
    WHERE preference.principal_id = p_principal_id
      AND preference.surface = p_surface;

    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6102',
            MESSAGE = 'analytics dashboard default was not found';
    END IF;

    RETURN QUERY
    SELECT *
    FROM analytics_dashboard_get_v1(p_scope_schema, v_dashboard_id);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_get_default_v1(name,uuid,text)
    FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_get_default_v1(name,uuid,text)
    TO harpy_runtime;
