-- +goose Up
GRANT EXECUTE ON FUNCTION iam_permission_request_audit_write_v1(
    name,uuid,bytea,uuid,uuid,uuid,text,text,bigint,bigint
) TO forgecrud_runtime;
