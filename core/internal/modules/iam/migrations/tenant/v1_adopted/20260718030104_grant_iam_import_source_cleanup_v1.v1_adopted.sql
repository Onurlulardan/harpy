-- +goose Up
GRANT EXECUTE ON FUNCTION iam_system_import_source_cleanup_claim_v1(
    name,uuid,uuid,timestamptz
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_system_import_source_cleanup_complete_v1(
    name,uuid,uuid,uuid,text,text,text,uuid,bigint,text,bytea,bytea,timestamptz
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_system_import_source_cleanup_retry_v1(
    name,uuid,uuid,uuid,bigint,text,timestamptz,timestamptz
) TO harpy_runtime;
