-- +goose Up
REVOKE EXECUTE ON FUNCTION iam_maintenance_profile_resources_read_v1(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer
) FROM forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_profile_resources_read_v2(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer
) TO forgecrud_runtime;

