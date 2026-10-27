-- +goose Up
GRANT EXECUTE ON FUNCTION iam_profile_publication_receipt_read_v1(
    name,uuid,uuid,uuid,bigint,bigint
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_profile_backfill_source_resources_plan_v1(
    name,text,uuid,uuid,integer,uuid,uuid,bigint,bigint
) TO forgecrud_runtime;

