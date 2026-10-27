-- +goose Up
REVOKE EXECUTE ON FUNCTION iam_current_session_read_v4(name,uuid,uuid)
FROM forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_current_session_read_v4(name,uuid,uuid)
TO forgecrud_runtime;
