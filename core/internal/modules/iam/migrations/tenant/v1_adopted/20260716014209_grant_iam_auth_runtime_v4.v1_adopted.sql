-- IAM authentication runtime ACL boundary.
-- +goose Up
REVOKE EXECUTE ON FUNCTION iam_challenge_resolve_v1(
    name,uuid,bigint,smallint,text
) FROM harpy_runtime;

REVOKE EXECUTE ON FUNCTION iam_recovery_consume_v1(
    name,uuid,bigint
) FROM harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_reauthentication_start_v1(
    name,uuid,bytea,uuid,uuid,uuid,bigint,bigint,text,uuid,text,uuid,text,
    bytea,bytea,bytea,bigint,bytea,boolean
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_reauthentication_credential_v1(
    name,uuid,bytea,uuid,uuid,uuid,bigint,bigint,text,uuid,text,bytea,
    bigint,bytea,bytea
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_reauthentication_reject_v1(
    name,uuid,bigint,smallint,uuid,bigint,boolean,boolean
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_reauthentication_complete_v1(
    name,uuid,bytea,uuid,bigint,smallint,uuid,uuid,bigint,bigint,bigint,
    bigint,uuid,text,bigint,uuid,bigint,uuid,uuid,bigint,bigint,uuid,
    bigint,bigint,bigint,bigint,bytea,bytea,text,bigint,boolean,boolean,
    bigint,uuid,bytea,text,bytea
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_session_revoke_v1(
    name,uuid,uuid,bigint,bigint,uuid,bigint,uuid,bigint
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_recovery_credential_v1(
    name,bytea,text,bytea
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_recovery_consume_v2(
    name,bytea,text,bytea,uuid,bigint,uuid,bigint,uuid,bigint,bigint
) TO harpy_runtime;
