-- +goose Up
REVOKE EXECUTE ON FUNCTION iam_refresh_rotate_v1(
    name,uuid,bigint,uuid,bytea,text,bytea
) FROM harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_login_credential_v1(
    name,uuid,uuid,text,bigint,bytea,text
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_password_use_v1(
    name,uuid,bigint,bigint,bigint,bigint,bigint,boolean,
    bytea,bytea,integer,smallint,smallint
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_session_issue_v1(
    name,uuid,bigint,bigint,bigint,uuid,text,bigint,uuid,bigint,
    uuid,text,text,text,text,bytea,uuid,uuid,bytea,text,bytea
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_refresh_credential_v1(
    name,uuid,uuid,bytea,bigint,bytea
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_refresh_rotate_v2(
    name,uuid,uuid,bigint,bigint,uuid,bigint,uuid,bigint,bigint,bigint,
    uuid,bigint,uuid,bigint,uuid,bytea,text,bytea
) TO harpy_runtime;
