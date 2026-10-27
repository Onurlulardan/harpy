-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_resource_manifest_v1(
    p_scope_schema name, p_resource_kind text, p_version_id uuid
)
RETURNS bytea
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE v_manifest bytea;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    IF p_resource_kind = 'query' THEN
        SELECT version_row.manifest_sha256 INTO v_manifest
        FROM analytics_saved_query_versions AS version_row
        WHERE version_row.id = p_version_id
          AND version_row.version_state = 'published';
    ELSIF p_resource_kind = 'dashboard' THEN
        SELECT version_row.manifest_sha256 INTO v_manifest
        FROM analytics_dashboard_versions AS version_row
        WHERE version_row.id = p_version_id
          AND version_row.version_state = 'published';
    ELSE
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics resource kind is invalid';
    END IF;
    IF pg_catalog.octet_length(v_manifest) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='analytics immutable resource was not found';
    END IF;
    RETURN v_manifest;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_resource_manifest_v1(name,text,uuid) FROM PUBLIC;
