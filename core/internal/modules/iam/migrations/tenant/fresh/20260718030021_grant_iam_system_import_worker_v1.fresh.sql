-- +goose Up

GRANT EXECUTE ON FUNCTION iam_system_record_import_worker_claim_v1(
    name,uuid,uuid,uuid,uuid,uuid,bigint
) TO forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_import_worker_receipts_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,bigint[]
) TO forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_import_worker_commit_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,jsonb,bigint
) TO forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_import_worker_finish_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,boolean
) TO forgecrud_v2_runtime;
