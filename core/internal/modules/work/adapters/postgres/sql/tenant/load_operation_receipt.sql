SELECT request_sha256, state, aggregate_id, COALESCE(aggregate_version, 0)
FROM work_operations
WHERE actor_id = $1 AND operation_kind = $2 AND idempotency_key = $3
FOR UPDATE
