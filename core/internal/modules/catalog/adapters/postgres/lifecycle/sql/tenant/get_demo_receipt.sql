SELECT receipt_id, installation_id, operation_id, component_id, action,
       stable_ids, content_sha256, actor_id, idempotency_key, completed_at
FROM catalog_demo_receipts
WHERE installation_id = $1::uuid AND component_id = $2::uuid
  AND action = $3::text AND idempotency_key = $4::text

