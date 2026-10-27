-- +goose Up
REVOKE EXECUTE ON FUNCTION iam_session_issue_v1(
    name,uuid,bigint,bigint,bigint,uuid,text,bigint,uuid,bigint,
    uuid,text,text,text,text,bytea,uuid,uuid,bytea,text,bytea
) FROM forgecrud_v2_runtime;
REVOKE EXECUTE ON FUNCTION iam_refresh_credential_v2(
    name,uuid,uuid,bytea,bigint,bytea
) FROM forgecrud_v2_runtime;
REVOKE EXECUTE ON FUNCTION iam_refresh_rotate_v3(
    name,uuid,uuid,bigint,bigint,uuid,bigint,uuid,bigint,bigint,bigint,
    uuid,bigint,uuid,bigint,uuid,bytea,text,bytea
) FROM forgecrud_v2_runtime;
REVOKE EXECUTE ON FUNCTION iam_current_session_read_v1(name,uuid,uuid)
FROM forgecrud_v2_runtime;

GRANT EXECUTE ON FUNCTION iam_security_epoch_current_v1(name)
TO forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_session_issue_v2(
    name,uuid,bigint,bigint,bigint,uuid,text,bigint,uuid,bigint,
    uuid,text,text,text,text,bytea,uuid,uuid,bytea,text,bytea
) TO forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_refresh_credential_v3(
    name,uuid,uuid,bytea,bigint,bytea
) TO forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_refresh_rotate_v4(
    name,uuid,uuid,bigint,bigint,uuid,bigint,uuid,bigint,bigint,bigint,
    uuid,bigint,uuid,bigint,uuid,bytea,text,bytea,bigint,bytea
) TO forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_current_session_read_v2(name,uuid,uuid)
TO forgecrud_v2_runtime;
