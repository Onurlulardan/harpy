-- +goose Up
REVOKE EXECUTE ON FUNCTION iam_refresh_credential_v1(
    name,uuid,uuid,bytea,bigint,bytea
) FROM forgecrud_runtime;

REVOKE EXECUTE ON FUNCTION iam_refresh_rotate_v2(
    name,uuid,uuid,bigint,bigint,uuid,bigint,uuid,bigint,bigint,bigint,
    uuid,bigint,uuid,bigint,uuid,bytea,text,bytea
) FROM forgecrud_runtime;

GRANT EXECUTE ON FUNCTION iam_refresh_credential_v2(
    name,uuid,uuid,bytea,bigint,bytea
) TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION iam_refresh_rotate_v3(
    name,uuid,uuid,bigint,bigint,uuid,bigint,uuid,bigint,bigint,bigint,
    uuid,bigint,uuid,bigint,uuid,bytea,text,bytea
) TO forgecrud_runtime;
