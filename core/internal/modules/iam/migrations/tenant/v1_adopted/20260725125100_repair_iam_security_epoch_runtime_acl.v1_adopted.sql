-- +goose Up
REVOKE EXECUTE ON FUNCTION iam_security_epoch_current_v1(name)
FROM harpy_v2_runtime;
REVOKE EXECUTE ON FUNCTION iam_session_issue_v2(
    name,uuid,bigint,bigint,bigint,uuid,text,bigint,uuid,bigint,
    uuid,text,text,text,text,bytea,uuid,uuid,bytea,text,bytea
) FROM harpy_v2_runtime;
REVOKE EXECUTE ON FUNCTION iam_refresh_credential_v3(
    name,uuid,uuid,bytea,bigint,bytea
) FROM harpy_v2_runtime;
REVOKE EXECUTE ON FUNCTION iam_refresh_rotate_v4(
    name,uuid,uuid,bigint,bigint,uuid,bigint,uuid,bigint,bigint,bigint,
    uuid,bigint,uuid,bigint,uuid,bytea,text,bytea,bigint,bytea
) FROM harpy_v2_runtime;
REVOKE EXECUTE ON FUNCTION iam_current_session_read_v2(name,uuid,uuid)
FROM harpy_v2_runtime;
REVOKE EXECUTE ON FUNCTION iam_current_session_read_v3(name,uuid,uuid)
FROM harpy_v2_runtime;
REVOKE EXECUTE ON FUNCTION iam_security_epoch_assert_v1(name,bytea)
FROM harpy_v2_runtime;

REVOKE EXECUTE ON FUNCTION iam_session_issue_v1(
    name,uuid,bigint,bigint,bigint,uuid,text,bigint,uuid,bigint,
    uuid,text,text,text,text,bytea,uuid,uuid,bytea,text,bytea
) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION iam_refresh_credential_v2(
    name,uuid,uuid,bytea,bigint,bytea
) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION iam_refresh_rotate_v3(
    name,uuid,uuid,bigint,bigint,uuid,bigint,uuid,bigint,bigint,bigint,
    uuid,bigint,uuid,bigint,uuid,bytea,text,bytea
) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION iam_current_session_read_v1(name,uuid,uuid)
FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION iam_current_session_read_v2(name,uuid,uuid)
FROM harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_security_epoch_current_v1(name)
TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_session_issue_v2(
    name,uuid,bigint,bigint,bigint,uuid,text,bigint,uuid,bigint,
    uuid,text,text,text,text,bytea,uuid,uuid,bytea,text,bytea
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_refresh_credential_v3(
    name,uuid,uuid,bytea,bigint,bytea
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_refresh_rotate_v4(
    name,uuid,uuid,bigint,bigint,uuid,bigint,uuid,bigint,bigint,bigint,
    uuid,bigint,uuid,bigint,uuid,bytea,text,bytea,bigint,bytea
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_current_session_read_v3(name,uuid,uuid)
TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_security_epoch_assert_v1(name,bytea)
TO harpy_runtime;
