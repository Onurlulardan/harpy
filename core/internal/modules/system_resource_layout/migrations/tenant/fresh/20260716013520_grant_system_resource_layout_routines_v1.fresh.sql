-- +goose Up
GRANT EXECUTE ON FUNCTION system_resource_layout_create_v1(
    name,uuid,text,text,bytea,timestamptz
) TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION system_resource_layout_get_v1(name,uuid)
    TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION system_resource_layout_list_v1(name,text)
    TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION system_resource_layout_update_v1(
    name,uuid,bigint,text,bytea,timestamptz
) TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION system_resource_layout_delete_v1(
    name,uuid,bigint
) TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION system_resource_layout_publish_v1(
    name,uuid,bigint,text,timestamptz
) TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION system_resource_layout_resolve_v1(name,text)
    TO forgecrud_runtime;

