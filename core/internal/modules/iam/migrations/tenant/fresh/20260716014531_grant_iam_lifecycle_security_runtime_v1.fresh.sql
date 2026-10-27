-- +goose Up
GRANT EXECUTE ON FUNCTION iam_totp_enrollment_start_v1(
    name,uuid,uuid,bytea,uuid,uuid,uuid,uuid,text,text,bytea,bytea,bytea,bytea,
    text,bytea,bytea,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_totp_enrollment_read_v1(
    name,uuid,uuid,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_totp_enrollment_complete_v1(
    name,uuid,uuid,bytea,uuid,uuid,text,bytea,bytea,bigint,bigint,bigint,
    uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_recovery_codes_issue_v1(
    name,uuid,uuid,bytea,uuid,uuid,text,jsonb,text,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_recovery_codes_list_v1(
    name,uuid,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_session_list_v1(
    name,uuid,uuid,uuid,bigint,bigint,integer,timestamptz,uuid
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_session_revoke_one_v1(
    name,uuid,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_session_revoke_all_v1(
    name,uuid,uuid,bytea,uuid,uuid,bigint,bigint
) TO harpy_runtime;
