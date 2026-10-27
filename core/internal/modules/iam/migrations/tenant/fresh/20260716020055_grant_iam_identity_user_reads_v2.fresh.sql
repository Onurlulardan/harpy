-- +goose Up
GRANT EXECUTE ON FUNCTION iam_identity_user_read_v2(
    name,uuid,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_user_list_v2(
    name,uuid,integer,uuid,uuid,bigint,bigint
) TO harpy_runtime;

REVOKE EXECUTE ON FUNCTION iam_identity_user_read_v1(
    name,uuid,uuid,uuid,bigint,bigint
) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION iam_identity_user_list_v1(
    name,uuid,integer,uuid,uuid,bigint,bigint
) FROM harpy_runtime;
