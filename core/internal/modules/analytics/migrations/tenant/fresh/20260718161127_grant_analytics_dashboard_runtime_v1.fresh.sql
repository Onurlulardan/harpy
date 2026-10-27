-- +goose Up

GRANT EXECUTE ON FUNCTION analytics_dashboard_get_v1(name,uuid) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_get_version_v1(name,uuid,uuid) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_list_v1(name,integer,uuid) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_create_v1(
    name,uuid,uuid,text,bytea,uuid,uuid,text,text,bytea,jsonb,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_update_v1(
    name,uuid,uuid,text,bytea,uuid,bigint,bigint,text,text,bytea,jsonb,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_delete_v1(
    name,uuid,uuid,text,bytea,uuid,bigint,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_publish_v1(
    name,uuid,uuid,text,bytea,uuid,uuid,uuid,bigint,bigint,bytea,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_grant_acl_v1(
    name,uuid,uuid,text,bytea,uuid,text,uuid,text,bigint,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_revoke_acl_v1(
    name,uuid,uuid,text,bytea,uuid,text,uuid,bigint,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_set_default_v1(
    name,uuid,uuid,text,bytea,uuid,uuid,text,timestamptz
) TO forgecrud_runtime;
