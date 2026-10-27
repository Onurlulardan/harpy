-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_get_v1(
    p_scope_schema name,
    p_revision_id uuid
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
       OR p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'system resource layout get scope is invalid';
    END IF;
    v_document := system_resource_layout_revision_json_v1(p_revision_id);
    IF v_document IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'system resource layout revision was not found';
    END IF;
    RETURN v_document;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_get_v1(name,uuid)
    FROM PUBLIC;
