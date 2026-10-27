WITH lock_attempt AS (
    SELECT pg_catalog.pg_try_advisory_xact_lock($1::bigint) AS locked
)
SELECT
    lock_attempt.locked,
    receipt.receipt_id::text,
    receipt.request_sha256,
    receipt.authorization_version,
    receipt.retention_class,
    receipt.response_status,
    receipt.response_headers,
    receipt.response_body,
    receipt.http_result_ref,
    receipt.response_sha256,
    receipt.async_job_id::text,
    receipt.completed_at,
    receipt.retain_until
FROM lock_attempt
LEFT JOIN LATERAL (
    SELECT candidate.*
    FROM platform.idempotency_receipts AS candidate
    WHERE lock_attempt.locked
      AND candidate.record_kind = 'http'
      AND candidate.tenant_id = $2
      AND candidate.principal_id = $3
      AND candidate.operation_id = $4
      AND candidate.idempotency_key = $5
) AS receipt ON true;
