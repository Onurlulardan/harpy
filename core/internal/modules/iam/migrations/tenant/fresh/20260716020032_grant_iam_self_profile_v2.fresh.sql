-- +goose Up

GRANT EXECUTE ON FUNCTION iam_self_profile_update_v2(
    name,uuid,bytea,uuid,uuid,uuid,bytea,bytea,bigint,bigint,bigint
) TO harpy_runtime;
