-- +goose Up
GRANT EXECUTE ON FUNCTION platform.iam_resolve_tenant_target_v1(uuid)
TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION platform.iam_resolve_login_target_v1(uuid,text,text,bytea)
TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION platform.iam_resolve_refresh_target_v1(bytea)
TO forgecrud_runtime;
