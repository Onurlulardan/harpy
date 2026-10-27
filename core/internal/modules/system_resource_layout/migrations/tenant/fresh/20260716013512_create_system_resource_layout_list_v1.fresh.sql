-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_list_v1(
    p_scope_schema name,
    p_resource_kind text
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_document jsonb;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR p_scope_schema IS DISTINCT FROM pg_catalog.current_schema()
       OR p_resource_kind NOT IN ('user', 'organization', 'role') THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'system resource layout list scope is invalid';
    END IF;
    SELECT COALESCE(
        pg_catalog.jsonb_agg(
            system_resource_layout_revision_json_v1(revision.id)
            ORDER BY revision.version
        ),
        '[]'::jsonb
    )
    INTO v_document
    FROM system_resource_layout_revisions AS revision
    WHERE revision.resource_kind = p_resource_kind;
    RETURN v_document;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_list_v1(name,text)
    FROM PUBLIC;
