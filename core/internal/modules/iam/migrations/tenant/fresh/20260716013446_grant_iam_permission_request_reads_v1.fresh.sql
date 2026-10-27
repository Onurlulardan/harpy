-- +goose Up
GRANT EXECUTE ON FUNCTION iam_permission_request_read_v1(
    name,uuid,uuid
) TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION iam_permission_request_list_v1(
    name,uuid,jsonb,uuid,integer
) TO forgecrud_runtime;
