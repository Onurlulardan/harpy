-- +goose Up

GRANT EXECUTE ON FUNCTION analytics_query_get_v1(name,uuid)
    TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_query_get_version_v1(name,uuid,uuid)
    TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_query_list_v1(name,integer,uuid)
    TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_query_create_v1(
    name,uuid,uuid,text,bytea,uuid,uuid,text,bytea,bytea,bytea,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_query_update_v1(
    name,uuid,uuid,text,bytea,uuid,bigint,bigint,text,bytea,bytea,bytea,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_query_delete_v1(
    name,uuid,uuid,text,bytea,uuid,bigint,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_query_publish_v1(
    name,uuid,uuid,text,bytea,uuid,uuid,uuid,bigint,bigint,
    bytea,bytea,bytea,jsonb,bytea,text,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_query_revalidate_v1(
    name,uuid,uuid,text,bytea,uuid,uuid,uuid,bigint,bigint,
    bytea,bytea,bytea,jsonb,bytea,text,timestamptz
) TO forgecrud_runtime;
