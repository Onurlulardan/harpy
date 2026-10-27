-- +goose Up
GRANT EXECUTE ON FUNCTION iam_lifecycle_replay_v1(
    uuid,text,bytea,uuid
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_lifecycle_record_v1(
    uuid,text,bytea,uuid,uuid,bigint,text,uuid,text,boolean,boolean
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_lifecycle_event_sequence_v1(
    uuid,text,text,uuid
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identifier_add_v1(
    name,uuid,uuid,bytea,uuid,uuid,text,text,text,bytea,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identifier_verify_v2(
    name,uuid,uuid,bytea,bytea,text,bytea,bytea
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identifier_tombstone_v1(
    name,uuid,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_identifier_retire_v1(
    name,uuid,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_one_time_issue_v1(
    uuid,text,uuid,uuid,text,bigint,uuid,bigint,bytea,text,bytea,bytea
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_email_verification_start_v1(
    name,uuid,uuid,bytea,uuid,uuid,bytea,text,bytea,bytea,
    uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_password_reset_start_v1(
    name,uuid,uuid,bytea,uuid,text,bytea,bytea,text,bytea,bytea
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_password_reset_complete_v1(
    name,uuid,uuid,bytea,bytea,text,bytea,bytea,bytea,bytea
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_portal_setup_start_v1(
    name,uuid,uuid,bytea,uuid,uuid,bytea,text,bytea,bytea,
    uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_portal_setup_complete_v2(
    name,uuid,uuid,bytea,bytea,text,bytea,bytea,bytea,bytea
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_portal_approval_decide_v1(
    name,uuid,uuid,bytea,uuid,bigint,text,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_password_change_v1(
    name,uuid,uuid,bytea,uuid,bigint,bytea,bytea,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_authenticator_list_v1(
    name,uuid,uuid,uuid,bigint,bigint,integer,timestamptz,uuid
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_authenticator_suspend_v1(
    name,uuid,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_authenticator_retire_v1(
    name,uuid,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_webauthn_enrollment_start_v1(
    name,uuid,uuid,bytea,uuid,uuid,uuid,uuid,text,text,bytea,bytea,bigint,
    bytea,bytea,boolean,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_webauthn_enrollment_complete_v1(
    name,uuid,uuid,bytea,uuid,uuid,text,bytea,bytea,bytea,bytea,text,bytea,
    bigint,uuid,bigint,boolean,boolean,boolean,boolean,uuid,uuid,bigint,bigint
) TO harpy_runtime;
