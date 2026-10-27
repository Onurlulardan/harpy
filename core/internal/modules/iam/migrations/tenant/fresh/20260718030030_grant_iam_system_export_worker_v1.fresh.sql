-- +goose Up
GRANT EXECUTE ON FUNCTION iam_system_record_export_worker_load_v1(
    name,uuid,uuid,uuid,text
) TO forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_worker_claim_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint
) TO forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_worker_checkpoint_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,
    bigint,bigint,timestamptz
) TO forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_worker_effect_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,
    uuid,uuid,text,text,bytea,bytea,bigint,bytea,bytea,bytea,timestamptz
) TO forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_worker_complete_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,bigint,
    bigint,timestamptz,text,text,text,text,uuid,uuid,bytea,bytea,bytea,bytea,
    bigint,bytea,bytea,timestamptz
) TO forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_worker_finish_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,text,text,
    timestamptz
) TO forgecrud_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_system_record_export_worker_unknown_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,uuid,uuid,
    text,text,bytea,bytea,bigint,bytea,bytea,bytea,timestamptz,boolean,text,
    text,bytea,timestamptz,timestamptz,text
) TO forgecrud_v2_runtime;
