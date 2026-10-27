-- +goose Up
GRANT EXECUTE ON FUNCTION iam_capability_grant_read_v1(name,uuid,bigint)
    TO harpy_runtime;
