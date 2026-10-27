-- +goose Up
GRANT EXECUTE ON FUNCTION platform.iam_post_restore_reconcile_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,bytea,bytea,bytea,bigint,bytea,jsonb,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.iam_post_restore_admission_v1(boolean)
TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.iam_post_restore_challenge_v1(
    uuid,uuid,uuid,text,bytea,text,text,timestamptz,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.iam_post_restore_challenge_context_v1(
    uuid,uuid,uuid,text,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.iam_post_restore_binding_v1(uuid,uuid)
TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.iam_post_restore_reactivation_fail_v1(
    uuid,uuid,uuid,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.iam_post_restore_reactivation_complete_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,bytea,bytea,bytea,bigint,timestamptz
) TO harpy_runtime;
