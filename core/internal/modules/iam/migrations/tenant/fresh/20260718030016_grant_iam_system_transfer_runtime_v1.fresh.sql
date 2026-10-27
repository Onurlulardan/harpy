-- +goose Up

GRANT EXECUTE ON FUNCTION iam_system_import_source_register_v1(
    name,uuid,uuid,uuid,text,bigint,bigint,bytea,text,text,text,bytea,bytea,
    bigint,bytea,bigint,text,bytea,bytea,text,timestamptz
) TO forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_submit_v1(
    name,uuid,uuid,uuid,uuid,text,bigint,bigint,bigint,uuid,bytea,bytea,
    bytea,bytea,bytea,text,timestamptz
) TO forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_import_submit_v1(
    name,uuid,uuid,uuid,uuid,text,uuid,bigint,bigint,bigint,uuid,bytea,bytea,
    bytea,bytea,text,bytea,bytea,integer,timestamptz
) TO forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_read_v1(
    name,uuid,uuid,text,bigint,bigint
) TO forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_content_v1(
    name,uuid,uuid,text,bigint,bigint,timestamptz
) TO forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_cancel_v1(
    name,uuid,uuid,text,bigint,bigint,timestamptz
) TO forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_import_read_v1(
    name,uuid,uuid,text,bigint,bigint
) TO forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_import_cancel_v1(
    name,uuid,uuid,text,bigint,bigint,timestamptz
) TO forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_import_receipts_v1(
    name,uuid,uuid,text,bigint,bigint,bigint,integer
) TO forgecrud_v2_runtime;
