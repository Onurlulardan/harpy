-- +goose Up
GRANT EXECUTE ON FUNCTION iam_invitation_role_lookup_v1(
    name,uuid,text
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_invitation_credential_lookup_v1(
    name,bytea,bytea
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_invitation_read_v2(
    name,uuid,uuid
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_invitation_list_v2(
    name,uuid,jsonb,uuid,integer
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_invitation_register_accept_v1(
    name,uuid,bytea,uuid,bigint,uuid,bigint,bytea,text,
    bytea,bytea,uuid,uuid,uuid,uuid,uuid,uuid,bytea,bytea,bytea,bytea,bigint
) TO harpy_runtime;

REVOKE EXECUTE ON FUNCTION iam_invitation_read_v1(
    name,uuid,uuid
) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION iam_invitation_list_v1(
    name,uuid,jsonb,uuid,integer
) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION iam_invitation_accept_v3(
    name,uuid,bytea,uuid,bigint,uuid,bigint,bytea,text,bytea,bytea,
    uuid,uuid,uuid,uuid,bigint
) FROM harpy_runtime;
