-- +goose Up
GRANT EXECUTE ON FUNCTION iam_invitation_create_v2(
    name,uuid,bytea,uuid,text,text,bytea,uuid,text,uuid,uuid,uuid,bigint,
    uuid,bytea,text,bytea,bytea,uuid,text,text,bytea,bytea,bytea,bytea,bytea
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_invitation_revoke_v2(
    name,uuid,bytea,uuid,bigint,uuid,bigint
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_invitation_expire_v2(
    name,uuid,bytea,uuid,bigint,uuid,bigint
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_invitation_accept_v3(
    name,uuid,bytea,uuid,bigint,uuid,bigint,bytea,text,bytea,bytea,
    uuid,uuid,uuid,uuid,bigint
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_invitation_audit_write_v1(
    name,uuid,uuid,uuid,bytea,text,text,bigint,text,bigint
) TO forgecrud_runtime;

REVOKE EXECUTE ON FUNCTION iam_invitation_create_v1(
    name,uuid,text,text,bytea,uuid,text,uuid,uuid,uuid,bigint,
    uuid,bytea,text,bytea,bytea
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_invitation_revoke_v1(
    name,uuid,bigint,uuid,bigint
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_invitation_expire_v1(
    name,uuid,bigint,uuid,bigint
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_invitation_accept_v2(
    name,uuid,bytea,uuid,bigint,uuid,bigint,bytea,text,bytea,bytea,
    uuid,uuid,uuid,uuid,bigint
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_invitation_grant_role_v1(
    uuid,uuid,uuid,uuid,text,uuid,uuid,timestamptz
) FROM forgecrud_runtime;
