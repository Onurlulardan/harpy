-- +goose Up
GRANT EXECUTE ON FUNCTION iam_reconcile_restored_identity_v1(
    uuid,uuid,uuid,uuid,uuid,bytea,bytea,bytea,bigint,bytea,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_post_restore_owner_context_v1(uuid)
TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_post_restore_reactivate_owner_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,bytea,bytea,bytea,bigint,bytea,bytea,
    integer,smallint,smallint,uuid,bigint,text,bytea,timestamptz
) TO harpy_runtime;
