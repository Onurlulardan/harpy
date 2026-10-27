-- +goose Up

REVOKE EXECUTE ON FUNCTION iam_identity_organization_delete_v1(
    name,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) FROM forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_organization_delete_v2(
    name,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) TO forgecrud_runtime;
