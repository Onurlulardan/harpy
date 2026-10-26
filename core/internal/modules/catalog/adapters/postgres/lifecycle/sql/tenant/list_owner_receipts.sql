SELECT receipt_id, operation_id, component_id, owner_module, owner_generation,
       owner_contract, content_sha256, action, ready, completed_at
FROM catalog_operation_steps
WHERE operation_id = $1::uuid
ORDER BY owner_module, component_id, action
