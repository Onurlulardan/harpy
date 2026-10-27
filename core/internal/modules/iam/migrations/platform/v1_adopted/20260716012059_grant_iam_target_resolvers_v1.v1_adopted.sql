-- +goose Up
GRANT EXECUTE ON FUNCTION platform.iam_resolve_tenant_target_v1(uuid)
TO harpy_runtime;

GRANT EXECUTE ON FUNCTION platform.iam_resolve_login_target_v1(uuid,text,text,bytea)
TO harpy_runtime;

GRANT EXECUTE ON FUNCTION platform.iam_resolve_refresh_target_v1(bytea)
TO harpy_runtime;
