-- +goose Up
GRANT EXECUTE ON FUNCTION iam_permission_request_group_create_v1(
    name,uuid,uuid,text,uuid,uuid,bigint,bytea,timestamptz,jsonb
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_permission_request_review_v1(
    name,uuid,bigint,text,uuid,bigint,uuid
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_permission_request_cancel_v1(
    name,uuid,bigint,text,uuid,bigint
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_permission_request_expire_v1(
    name,uuid,bigint,text
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_permission_request_decide_v1(
    name,uuid,bytea,uuid,bigint,text,uuid,bigint,bigint,uuid,text,
    bytea,timestamptz,text,boolean,boolean,boolean
) TO harpy_runtime;
