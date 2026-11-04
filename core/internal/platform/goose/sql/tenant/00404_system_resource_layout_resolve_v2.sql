-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION system_resource_layout_resolve_v2(
    arg_1 name,
    arg_2 text
)
RETURNS bytea
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT NULL::bytea;
$$;

REVOKE ALL ON FUNCTION system_resource_layout_resolve_v2(name, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION system_resource_layout_resolve_v2(name, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION system_resource_layout_resolve_v2(name, text);
