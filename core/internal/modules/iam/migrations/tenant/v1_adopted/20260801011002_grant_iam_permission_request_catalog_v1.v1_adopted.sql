-- +goose Up
GRANT EXECUTE ON FUNCTION iam_permission_request_resource_catalog_v1(
    name,uuid,text,text,uuid,integer,text
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_permission_request_action_catalog_v1(
    name,uuid,text,uuid,integer
) TO harpy_runtime;

