-- +goose Up
REVOKE INSERT ON TABLE catalog_demo_operation_plans FROM forgecrud_runtime;
GRANT INSERT (
    operation_id,
    installation_id,
    action,
    component_id,
    source_run_id,
    record_ids,
    expected_plan_sha256,
    created_at
) ON TABLE catalog_demo_operation_plans TO forgecrud_runtime;

