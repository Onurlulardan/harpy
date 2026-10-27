-- +goose Up
GRANT EXECUTE ON FUNCTION platform.iam_auth_throttle_blocked_v2(
    text,uuid,text,text,text,bytea,bigint,bigint
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.iam_auth_throttle_record_failure_v2(
    text,uuid,text,text,text,bytea,bigint,bigint
) TO forgecrud_runtime;
