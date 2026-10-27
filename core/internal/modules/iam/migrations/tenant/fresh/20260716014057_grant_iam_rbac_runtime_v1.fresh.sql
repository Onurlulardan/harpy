-- +goose Up
GRANT EXECUTE ON FUNCTION iam_rbac_resource_create_v1(
    name,uuid,bytea,uuid,text,text,text,text,text,boolean,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_resource_get_v1(name,uuid,uuid,uuid) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_resource_list_v1(name,text,integer,uuid,uuid) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_resource_update_v1(
    name,uuid,bytea,uuid,text,text,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_resource_retire_v1(
    name,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_rbac_action_create_v1(
    name,uuid,bytea,uuid,text,text,text,boolean,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_action_get_v1(name,uuid,uuid,uuid) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_action_list_v1(name,text,integer,uuid,uuid) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_action_update_v1(
    name,uuid,bytea,uuid,text,text,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_action_retire_v1(
    name,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_rbac_permission_create_v1(
    name,uuid,bytea,uuid,text,uuid,uuid,boolean,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_permission_get_v1(name,uuid,uuid,uuid) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_permission_list_v1(name,text,integer,uuid,uuid) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_permission_retire_v1(
    name,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_rbac_user_binding_grant_v1(
    name,uuid,bytea,uuid,uuid,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_user_binding_revoke_v1(
    name,uuid,bytea,uuid,uuid,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_user_binding_get_v1(name,uuid,uuid,uuid) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_user_binding_list_v1(
    name,uuid,text,integer,uuid,uuid
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_rbac_role_binding_grant_v1(
    name,uuid,bytea,uuid,uuid,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_role_binding_revoke_v1(
    name,uuid,bytea,uuid,uuid,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_role_binding_get_v1(name,uuid,uuid,uuid) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_role_binding_list_v1(
    name,uuid,text,integer,uuid,uuid
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_rbac_organization_binding_grant_v1(
    name,uuid,bytea,uuid,uuid,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_organization_binding_revoke_v1(
    name,uuid,bytea,uuid,uuid,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_organization_binding_get_v1(name,uuid,uuid,uuid) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_organization_binding_list_v1(
    name,uuid,text,integer,uuid,uuid
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_rbac_implication_create_v1(
    name,uuid,bytea,uuid,uuid,uuid,text,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_implication_get_v1(name,uuid,uuid,uuid) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_implication_list_v1(name,text,integer,uuid,uuid) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_implication_record_hit_v1(
    name,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_implication_retire_v1(
    name,uuid,bytea,uuid,bigint,uuid,uuid,uuid,bigint,bigint
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_rbac_protected_admin_get_v1(name,uuid,uuid,uuid) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_protected_admin_update_v1(
    name,uuid,bytea,uuid,integer,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_rbac_audit_write_v1(
    name,uuid,uuid,uuid,bytea,text,text,bigint,bigint
) TO harpy_runtime;
