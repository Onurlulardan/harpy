-- +goose Up
GRANT EXECUTE ON FUNCTION iam_rbac_assignment_create_v1(
    name,uuid,bytea,uuid,uuid,text,uuid,jsonb,uuid,uuid,bigint,bigint,
    bytea,timestamptz,boolean,boolean,boolean
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_rbac_assignment_update_v1(
    name,uuid,bytea,uuid,uuid,text,uuid,jsonb,bigint,uuid,uuid,bigint,bigint,
    bytea,timestamptz,boolean,boolean,boolean
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_rbac_assignment_retire_v1(
    name,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint,
    bytea,timestamptz,boolean,boolean,boolean
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_rbac_assignment_read_v1(
    name,uuid,uuid,uuid
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_rbac_assignment_list_v1(
    name,text,uuid,integer,uuid,uuid
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_rbac_effective_permissions_v1(
    name,text,uuid,integer,uuid,uuid
) TO harpy_runtime;
