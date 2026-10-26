INSERT INTO catalog_demo_operation_plans (
    operation_id, installation_id, action, component_id, source_run_id,
    record_ids, expected_plan_sha256, created_at
) VALUES (
    $1::uuid, $2::uuid, $3::text, $4::uuid, $5::uuid,
    $6::jsonb, $7::text, $8::timestamptz
)
ON CONFLICT (operation_id, component_id) DO NOTHING

