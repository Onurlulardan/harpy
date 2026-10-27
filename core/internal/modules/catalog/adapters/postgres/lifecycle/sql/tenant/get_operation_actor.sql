SELECT actor_id
FROM catalog_operation_runs
WHERE operation_id = $1::uuid
