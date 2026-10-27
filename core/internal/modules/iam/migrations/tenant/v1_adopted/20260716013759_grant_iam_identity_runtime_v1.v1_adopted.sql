-- +goose Up
GRANT EXECUTE ON FUNCTION iam_identity_user_create_v1(
    name,uuid,uuid,bytea,uuid,uuid,uuid,uuid,uuid,bytea,bytea,
    text,text,bytea,text,bytea,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_user_read_v1(
    name,uuid,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_user_list_v1(
    name,uuid,integer,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_user_update_v1(
    name,uuid,bytea,uuid,bigint,uuid,bytea,bytea,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_user_disable_v1(
    name,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_user_delete_v1(
    name,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_identity_organization_create_v1(
    name,uuid,bytea,uuid,text,text,uuid,uuid,bytea,bytea,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_organization_read_v1(
    name,uuid,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_organization_list_v1(
    name,uuid,integer,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_organization_update_v1(
    name,uuid,bytea,uuid,bigint,text,text,uuid,uuid,bytea,bytea,
    uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_organization_disable_v1(
    name,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_organization_delete_v1(
    name,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_identity_role_create_v1(
    name,uuid,uuid,bytea,uuid,text,uuid,uuid,text,text,boolean,
    uuid,bytea,bytea,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_role_read_v1(
    name,uuid,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_role_list_v1(
    name,uuid,integer,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_role_update_v1(
    name,uuid,bytea,uuid,bigint,text,text,boolean,uuid,bytea,bytea,
    uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_role_retire_v1(
    name,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_identity_membership_add_v1(
    name,uuid,bytea,uuid,uuid,uuid,text,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_membership_read_v1(
    name,uuid,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_membership_list_v1(
    name,uuid,integer,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_membership_update_v1(
    name,uuid,bytea,uuid,bigint,text,text,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_membership_delete_v1(
    name,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_identity_assignment_grant_v1(
    name,uuid,bytea,uuid,uuid,uuid,text,uuid,uuid,uuid,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_assignment_read_v1(
    name,uuid,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_assignment_list_v1(
    name,uuid,integer,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_assignment_update_v1(
    name,uuid,bytea,uuid,bigint,text,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identity_assignment_revoke_v1(
    name,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;
