-- +goose Up
GRANT EXECUTE ON FUNCTION iam_current_session_read_v1(
    name,uuid,uuid
) TO harpy_runtime;
