-- +goose Up
GRANT EXECUTE ON FUNCTION iam_self_profile_read_v1(
    name,uuid,bigint,bigint
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_self_profile_update_v1(
    name,uuid,bytea,uuid,uuid,uuid,bytea,bytea,bigint,bigint,bigint
) TO harpy_runtime;
