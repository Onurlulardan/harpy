-- +goose Up

GRANT EXECUTE ON FUNCTION iam_system_import_source_register_v1(
    name,uuid,uuid,uuid,text,bigint,bigint,bytea,text,text,text,bytea,bytea,
    bigint,bytea,bigint,text,bytea,bytea,text,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_import_source_register_v2(
    name,uuid,uuid,uuid,text,bigint,bigint,bytea,text,text,text,bytea,bytea,
    bigint,bytea,bigint,text,bytea,bytea,text,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_submit_v1(
    name,uuid,uuid,uuid,uuid,text,bigint,bigint,bigint,uuid,bytea,bytea,
    bytea,bytea,bytea,text,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_import_submit_v1(
    name,uuid,uuid,uuid,uuid,text,uuid,bigint,bigint,bigint,uuid,bytea,bytea,
    bytea,bytea,text,bytea,bytea,integer,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_read_v1(
    name,uuid,uuid,text,bigint,bigint
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_content_v1(
    name,uuid,uuid,text,bigint,bigint,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_cancel_v1(
    name,uuid,uuid,text,bigint,bigint,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_import_read_v1(
    name,uuid,uuid,text,bigint,bigint
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_import_cancel_v1(
    name,uuid,uuid,text,bigint,bigint,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_import_receipts_v1(
    name,uuid,uuid,text,bigint,bigint,bigint,integer
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_import_worker_claim_v1(
    name,uuid,uuid,uuid,uuid,uuid,bigint
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_import_worker_receipts_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,bigint[]
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_import_worker_commit_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,jsonb,bigint
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_import_worker_finish_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,boolean
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_worker_load_v1(
    name,uuid,uuid,uuid,text
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_worker_claim_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_worker_checkpoint_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,
    bigint,bigint,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_worker_effect_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,
    uuid,uuid,text,text,bytea,bytea,bigint,bytea,bytea,bytea,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_worker_complete_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,bigint,
    bigint,timestamptz,text,text,text,text,uuid,uuid,bytea,bytea,bytea,bytea,
    bigint,bytea,bytea,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_worker_finish_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,text,text,
    timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_worker_unknown_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,uuid,uuid,
    text,text,bytea,bytea,bigint,bytea,bytea,bytea,timestamptz,boolean,text,
    text,bytea,timestamptz,timestamptz,text
) TO forgecrud_runtime;
