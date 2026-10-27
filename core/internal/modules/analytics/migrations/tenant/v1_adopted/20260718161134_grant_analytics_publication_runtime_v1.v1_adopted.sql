-- +goose Up

GRANT EXECUTE ON FUNCTION analytics_publication_get_v1(name,uuid) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION analytics_publication_list_v1(name,integer,uuid) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION analytics_publication_create_v1(
    name,uuid,uuid,text,bytea,uuid,text,uuid,text,text,bytea,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION analytics_publication_submit_v1(
    name,uuid,uuid,text,bytea,uuid,bigint,timestamptz,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION analytics_publication_approve_v1(
    name,uuid,uuid,text,bytea,uuid,bigint,timestamptz,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION analytics_publication_activate_v1(
    name,uuid,uuid,text,bytea,uuid,bigint,timestamptz,bytea,bytea,timestamptz,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION analytics_publication_revoke_v1(
    name,uuid,uuid,text,bytea,uuid,bigint,timestamptz
) TO harpy_runtime;
