-- +goose Up

GRANT EXECUTE ON FUNCTION iam_identity_user_create_v2(
    name,uuid,uuid,bytea,uuid,uuid,uuid,uuid,uuid,bytea,bytea,
    text,text,bytea,text,bytea,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_user_update_v2(
    name,uuid,bytea,uuid,bigint,uuid,bytea,bytea,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_organization_create_v2(
    name,uuid,bytea,uuid,text,text,uuid,uuid,bytea,bytea,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_organization_update_v2(
    name,uuid,bytea,uuid,bigint,text,text,uuid,uuid,bytea,bytea,
    uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_role_create_v2(
    name,uuid,uuid,bytea,uuid,text,uuid,uuid,text,text,boolean,
    uuid,bytea,bytea,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_role_update_v2(
    name,uuid,bytea,uuid,bigint,text,text,boolean,uuid,bytea,bytea,
    uuid,uuid,bigint,bigint
) TO harpy_runtime;
