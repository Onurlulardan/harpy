-- +goose Up
REVOKE EXECUTE ON FUNCTION system_resource_layout_create_v2(
    name,uuid,text,text,bytea,timestamptz
) FROM forgecrud_runtime;

REVOKE EXECUTE ON FUNCTION system_resource_layout_publish_v4(
    name,uuid,bigint,bytea,text,text,text,bytea,text,text,bytea,timestamptz
) FROM forgecrud_runtime;

GRANT EXECUTE ON FUNCTION system_resource_layout_create_v3(
    name,uuid,text,text,bytea,timestamptz
) TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION system_resource_layout_publish_v5(
    name,uuid,bigint,bytea,text,text,text,bytea,text,text,bytea,timestamptz
) TO forgecrud_runtime;
