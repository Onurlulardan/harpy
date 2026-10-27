-- +goose Up
GRANT EXECUTE ON FUNCTION iam_workflow_permission_decide_v1(
    name,uuid,uuid,bytea,uuid,text,uuid,bigint,bigint,uuid,bigint,uuid,
    bytea,text,bytea,timestamptz
) TO forgecrud_runtime;
