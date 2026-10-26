SELECT operation_id, installation_id, action, component_id,
       source_run_id, record_ids, expected_plan_sha256, created_at
FROM catalog_demo_operation_plans
WHERE operation_id = $1::uuid
ORDER BY component_id

