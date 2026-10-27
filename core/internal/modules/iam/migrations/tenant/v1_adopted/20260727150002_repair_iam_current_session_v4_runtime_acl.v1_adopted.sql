-- +goose Up
REVOKE EXECUTE ON FUNCTION iam_current_session_read_v4(name,uuid,uuid)
FROM harpy_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_current_session_read_v4(name,uuid,uuid)
TO harpy_runtime;
