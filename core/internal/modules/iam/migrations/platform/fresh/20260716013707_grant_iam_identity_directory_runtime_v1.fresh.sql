-- +goose Up
GRANT EXECUTE ON FUNCTION platform.iam_identity_directory_upsert_v1(
    uuid,bigint,bigint,uuid,uuid,text,bytea,bigint,bytea
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.iam_identity_directory_tombstone_v1(
    uuid,bigint,bigint,uuid,bigint
) TO forgecrud_runtime;
