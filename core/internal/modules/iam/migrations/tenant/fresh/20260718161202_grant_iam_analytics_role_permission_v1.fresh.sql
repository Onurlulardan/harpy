-- +goose Up
GRANT EXECUTE ON FUNCTION iam_analytics_role_permission_v1(
    name,uuid,uuid,text,text
) TO forgecrud_runtime;
