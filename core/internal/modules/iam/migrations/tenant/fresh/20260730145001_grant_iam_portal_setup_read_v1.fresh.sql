-- +goose Up
GRANT EXECUTE ON FUNCTION iam_portal_setup_read_v1(
    name,uuid,uuid,uuid,uuid,bigint,bigint
) TO harpy_runtime;
