-- +goose Up
GRANT EXECUTE ON FUNCTION iam_system_record_export_reconcile_expired_v1(
    name,uuid
) TO forgecrud_runtime;
