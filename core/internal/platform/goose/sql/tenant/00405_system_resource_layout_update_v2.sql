-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION system_resource_layout_update_v2(
    arg_1 name,
    arg_2 uuid,
    arg_3 bigint,
    arg_4 text,
    arg_5 bytea,
    arg_6 timestamptz
)
RETURNS bytea
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT NULL::bytea;
$$;

REVOKE ALL ON FUNCTION system_resource_layout_update_v2(name, uuid, bigint, text, bytea, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION system_resource_layout_update_v2(name, uuid, bigint, text, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION system_resource_layout_update_v2(name, uuid, bigint, text, bytea, timestamptz);
