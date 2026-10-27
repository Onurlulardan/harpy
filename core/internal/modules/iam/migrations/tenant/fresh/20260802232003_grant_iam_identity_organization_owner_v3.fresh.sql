-- +goose Up

REVOKE EXECUTE ON FUNCTION iam_identity_organization_create_v1(
    name,uuid,bytea,uuid,text,text,uuid,uuid,bytea,bytea,uuid,uuid,bigint,bigint
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_identity_organization_update_v1(
    name,uuid,bytea,uuid,bigint,text,text,uuid,uuid,bytea,bytea,
    uuid,uuid,bigint,bigint
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_identity_organization_create_v2(
    name,uuid,bytea,uuid,text,text,uuid,uuid,bytea,bytea,uuid,uuid,bigint,bigint
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_identity_organization_update_v2(
    name,uuid,bytea,uuid,bigint,text,text,uuid,uuid,bytea,bytea,
    uuid,uuid,bigint,bigint
) FROM forgecrud_runtime;

GRANT EXECUTE ON FUNCTION iam_identity_organization_create_v3(
    name,uuid,bytea,uuid,text,text,uuid,uuid[],uuid[],uuid,bytea,bytea,
    uuid,uuid,bigint,bigint
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_organization_update_v3(
    name,uuid,bytea,uuid,bigint,text,text,uuid,boolean,uuid[],uuid[],
    uuid,bytea,bytea,uuid,uuid,bigint,bigint
) TO forgecrud_runtime;
