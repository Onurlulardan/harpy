-- +goose Up
GRANT EXECUTE ON FUNCTION iam_system_record_export_reconcile_claim_v1(
    name,uuid,uuid,uuid,uuid,uuid,bigint,text
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_reconcile_sealed_v1(
    name,uuid,uuid,uuid,uuid,uuid,bigint,text,uuid,uuid,bytea,timestamptz,
    text,text,text,text,uuid,uuid,bytea,bytea,bytea,bytea,bigint,bytea,bytea,
    bigint,timestamptz,text,bytea,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_reconcile_missing_v1(
    name,uuid,uuid,uuid,uuid,uuid,bigint,text,uuid,uuid,bytea,timestamptz,
    text,bytea,timestamptz,text
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_reconcile_retry_v1(
    name,uuid,uuid,uuid,uuid,uuid,bigint,text,uuid,uuid,bytea,timestamptz,
    bytea,timestamptz,timestamptz,text
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_reconcile_park_v1(
    name,uuid,uuid,uuid,uuid,uuid,bigint,text,uuid,uuid,bytea,timestamptz,
    bytea,timestamptz,text
) TO forgecrud_runtime;
