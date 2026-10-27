-- +goose Up

GRANT EXECUTE ON FUNCTION iam_invitation_credential_lookup_v2(
    name,bytea,bytea
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_invitation_register_accept_v2(
    name,uuid,bytea,uuid,bigint,uuid,bigint,bytea,text,
    bytea,bytea,uuid,uuid,uuid,uuid,uuid,uuid,bytea,bytea,bytea,bytea,bigint
) TO forgecrud_runtime;
