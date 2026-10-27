-- +goose Up
GRANT EXECUTE ON FUNCTION iam_rest_query_actor_versions_v1(name,uuid)
    TO forgecrud_runtime;

GRANT SELECT ON iam_rest_query_users_v1 TO forgecrud_runtime;
GRANT SELECT ON iam_rest_query_roles_v1 TO forgecrud_runtime;
GRANT SELECT ON iam_rest_query_organizations_v1 TO forgecrud_runtime;
