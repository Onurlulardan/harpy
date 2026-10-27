-- +goose Up
GRANT USAGE ON SCHEMA platform TO harpy_runtime;

GRANT EXECUTE ON FUNCTION platform.iam_refresh_locator_create_v1(
    bytea,uuid,uuid,uuid,bigint,bigint,bigint,bytea,text,
    timestamptz,timestamptz,timestamptz,timestamptz
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION platform.iam_refresh_locator_advance_v1(
    bytea,uuid,bigint,bigint,bigint,bigint,bytea,text,timestamptz,timestamptz
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION platform.iam_tenant_throttle_record_attempt_v1(
    uuid,text,text,text,bytea,bigint,bigint,integer,integer,integer,integer
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION platform.iam_environment_throttle_record_attempt_v1(
    text,text,text,bytea,integer,integer,integer,integer
) TO harpy_runtime;
