INSERT INTO work_operations (
    operation_id, operation_kind, actor_id, idempotency_key, request_sha256,
    aggregate_kind, state, aggregate_id, aggregate_version, response_status,
    response_document, occurred_at, completed_at
) VALUES ($1, $2, $3, $4, $5, $2, 'CLAIMED', NULL, NULL, NULL, NULL,
          transaction_timestamp(), NULL)
ON CONFLICT (actor_id, operation_kind, idempotency_key) DO NOTHING
