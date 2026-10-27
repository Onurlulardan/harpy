-- +goose Up
GRANT EXECUTE ON FUNCTION iam_invitation_create_v1(
    name,uuid,text,text,bytea,uuid,text,uuid,uuid,uuid,bigint,
    uuid,bytea,text,bytea,bytea
) TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION iam_invitation_read_v1(
    name,uuid,uuid
) TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION iam_invitation_list_v1(
    name,uuid,jsonb,uuid,integer
) TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION iam_invitation_revoke_v1(
    name,uuid,bigint,uuid,bigint
) TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION iam_invitation_expire_v1(
    name,uuid,bigint,uuid,bigint
) TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION iam_invitation_accept_v2(
    name,uuid,bytea,uuid,bigint,uuid,bigint,bytea,text,bytea,bytea,
    uuid,uuid,uuid,uuid,bigint
) TO forgecrud_runtime;
