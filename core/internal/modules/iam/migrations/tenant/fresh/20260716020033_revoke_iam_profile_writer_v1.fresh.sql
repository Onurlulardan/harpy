-- +goose Up

REVOKE EXECUTE ON FUNCTION iam_identity_user_create_v1(
    name,uuid,uuid,bytea,uuid,uuid,uuid,uuid,uuid,bytea,bytea,
    text,text,bytea,text,bytea,uuid,uuid,bigint,bigint
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_identity_user_update_v1(
    name,uuid,bytea,uuid,bigint,uuid,bytea,bytea,uuid,uuid,bigint,bigint
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_identity_organization_create_v1(
    name,uuid,bytea,uuid,text,text,uuid,uuid,bytea,bytea,uuid,uuid,bigint,bigint
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_identity_organization_update_v1(
    name,uuid,bytea,uuid,bigint,text,text,uuid,uuid,bytea,bytea,
    uuid,uuid,bigint,bigint
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_identity_role_create_v1(
    name,uuid,uuid,bytea,uuid,text,uuid,uuid,text,text,boolean,
    uuid,bytea,bytea,uuid,uuid,bigint,bigint
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_identity_role_update_v1(
    name,uuid,bytea,uuid,bigint,text,text,boolean,uuid,bytea,bytea,
    uuid,uuid,bigint,bigint
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_invitation_credential_lookup_v1(
    name,bytea,bytea
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_invitation_register_accept_v1(
    name,uuid,bytea,uuid,bigint,uuid,bigint,bytea,text,
    bytea,bytea,uuid,uuid,uuid,uuid,uuid,uuid,bytea,bytea,bytea,bytea,bigint
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_self_profile_update_v1(
    name,uuid,bytea,uuid,uuid,uuid,bytea,bytea,bigint,bigint,bigint
) FROM forgecrud_runtime;
