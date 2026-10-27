-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_query_root_by_version_v1(
    p_scope_schema name,
    p_version_id uuid
)
RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_query_id uuid;
BEGIN
    IF p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       pg_catalog.uuid_extract_version(p_version_id) IS DISTINCT FROM 7 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics query version identity is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT version_row.query_id
    INTO v_query_id
    FROM analytics_query_versions AS version_row
    JOIN analytics_saved_queries AS root
      ON root.id=version_row.query_id
    WHERE version_row.id=p_version_id
      AND version_row.version_state='published'
      AND root.lifecycle_state='active';
    IF NOT FOUND OR pg_catalog.uuid_extract_version(v_query_id) IS DISTINCT FROM 7 THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='analytics published query version was not found';
    END IF;
    RETURN v_query_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_query_root_by_version_v1(name,uuid)
FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_query_root_by_version_v1(name,uuid)
TO harpy_runtime;

