-- +goose Up
GRANT EXECUTE ON FUNCTION iam_refresh_rotate_v1(
    name,uuid,bigint,uuid,bytea,text,bytea
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_one_time_consume_v1(
    name,uuid,bigint,text
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_invitation_accept_v1(
    name,uuid,bigint,uuid,bigint,uuid,uuid
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_challenge_resolve_v1(
    name,uuid,bigint,smallint,text
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_recovery_consume_v1(
    name,uuid,bigint
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_portal_setup_complete_v1(
    name,uuid,bigint,uuid,bigint
) TO harpy_runtime;
