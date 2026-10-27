-- +goose Up
REVOKE EXECUTE ON FUNCTION platform.iam_resolve_login_target_v1(
    uuid,text,text,bytea
) FROM harpy_runtime;

REVOKE EXECUTE ON FUNCTION platform.iam_resolve_refresh_target_v1(bytea)
FROM harpy_runtime;

GRANT EXECUTE ON FUNCTION platform.iam_resolve_login_target_v2(
    uuid,text,text,bytea
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION platform.iam_resolve_refresh_target_v2(bytea)
TO harpy_runtime;
