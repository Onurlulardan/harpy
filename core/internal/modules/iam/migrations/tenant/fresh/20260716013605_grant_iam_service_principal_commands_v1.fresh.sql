-- +goose Up
GRANT EXECUTE ON FUNCTION iam_service_principal_create_v1(
    name,uuid,bytea,uuid,text,text,uuid,uuid,bigint,bigint,
    bytea,timestamptz,boolean,boolean,boolean
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_service_principal_state_v1(
    name,uuid,bytea,uuid,bigint,text,uuid,uuid,bigint,bigint,
    bytea,timestamptz,boolean,boolean,boolean
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_capability_grant_write_v1(
    name,uuid,bytea,uuid,uuid,bigint,text,text,text,text,uuid,
    timestamptz,timestamptz,bytea,uuid,uuid,bigint,bigint,
    bytea,timestamptz,boolean,boolean,boolean
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_capability_grant_revoke_v1(
    name,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint,
    bytea,timestamptz,boolean,boolean,boolean
) TO harpy_runtime;

