-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_resolve_v1(
    p_scope_schema name,
    p_resource_kind text
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_revision_id uuid;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR p_scope_schema IS DISTINCT FROM pg_catalog.current_schema()
       OR p_resource_kind NOT IN ('user', 'organization', 'role') THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'system resource layout resolve scope is invalid';
    END IF;
    SELECT publication.revision_id
    INTO v_revision_id
    FROM system_resource_layout_publications AS publication
    WHERE publication.resource_kind = p_resource_kind;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'published system resource layout was not found';
    END IF;
    RETURN system_resource_layout_revision_json_v1(v_revision_id);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_resolve_v1(name,text)
    FROM PUBLIC;
