-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_resolve_v2(
    p_scope_schema name,
    p_resource_kind text
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR p_scope_schema IS DISTINCT FROM pg_catalog.current_schema()
       OR p_resource_kind NOT IN ('user', 'organization', 'role') THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'system resource presentation resolve scope is invalid';
    END IF;
    RETURN system_resource_layout_publication_json_v2(p_resource_kind);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_resolve_v2(name,text)
    FROM PUBLIC;

REVOKE EXECUTE ON FUNCTION system_resource_layout_publish_v2(
    name,uuid,bigint,text,timestamptz
) FROM harpy_runtime;

REVOKE EXECUTE ON FUNCTION system_resource_layout_resolve_v1(name,text)
    FROM harpy_runtime;

GRANT EXECUTE ON FUNCTION system_resource_layout_publish_v3(
    name,uuid,bigint,bytea,text,text,text,bytea,text,text,bytea,timestamptz
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION system_resource_layout_resolve_v2(name,text)
    TO harpy_runtime;
