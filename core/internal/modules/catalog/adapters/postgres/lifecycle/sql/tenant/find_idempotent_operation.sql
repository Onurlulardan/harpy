SELECT operation_id, request_sha256
FROM catalog_operation_runs
WHERE actor_id = $1::uuid AND idempotency_key = $2::text
