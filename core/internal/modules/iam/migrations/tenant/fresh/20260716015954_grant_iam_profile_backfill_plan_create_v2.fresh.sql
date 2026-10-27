-- +goose Up
REVOKE EXECUTE ON FUNCTION iam_profile_backfill_plan_create_v1(
    name,uuid,uuid,bytea,bytea,bytea,uuid,uuid,bigint,bigint
) FROM forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_profile_backfill_plan_create_v2(
    name,uuid,uuid,bytea,bytea,bytea,uuid,uuid,bigint,bigint
) TO forgecrud_runtime;
