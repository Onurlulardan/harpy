-- +goose Up

GRANT EXECUTE ON FUNCTION iam_profile_recovery_parent_read_v1(
    name,uuid,uuid,uuid,bigint,bigint
) TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION iam_profile_recovery_plan_create_v1(
    name,uuid,uuid,bytea,bytea,bytea,uuid,uuid,bigint,bigint
) TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION iam_profile_recovery_receipt_read_v1(
    name,uuid,uuid,uuid,bigint,bigint
) TO forgecrud_runtime;
