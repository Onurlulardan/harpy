-- +goose Up
REVOKE EXECUTE ON FUNCTION system_resource_layout_create_v1(
    name,uuid,text,text,bytea,timestamptz
) FROM harpy_runtime;

REVOKE EXECUTE ON FUNCTION system_resource_layout_update_v1(
    name,uuid,bigint,text,bytea,timestamptz
) FROM harpy_runtime;

REVOKE EXECUTE ON FUNCTION system_resource_layout_publish_v3(
    name,uuid,bigint,bytea,text,text,text,bytea,text,text,bytea,timestamptz
) FROM harpy_runtime;

GRANT EXECUTE ON FUNCTION system_resource_layout_create_v2(
    name,uuid,text,text,bytea,timestamptz
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION system_resource_layout_update_v2(
    name,uuid,bigint,text,bytea,timestamptz
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION system_resource_layout_publish_v4(
    name,uuid,bigint,bytea,text,text,text,bytea,text,text,bytea,timestamptz
) TO harpy_runtime;
