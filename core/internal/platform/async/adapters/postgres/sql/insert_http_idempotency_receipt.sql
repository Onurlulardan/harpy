WITH db_time AS (
    SELECT pg_catalog.clock_timestamp() AS completed_at
), inserted AS (
    INSERT INTO platform.idempotency_receipts (
        receipt_id, record_kind, tenant_id, scope_kind, scope_id,
        authority_group_id, authority_group_generation,
        authority_domain, cutover_generation, writer_origin,
        request_sha256, retention_class, principal_id, operation_id,
        idempotency_key, authorization_version, response_status,
        response_headers, response_body, http_result_ref, response_sha256,
        async_job_id, completed_at, retain_until
    )
    SELECT
        $1, 'http', $2, 'tenant', $2,
        $3, $4, $5, $6, 'v2',
        $7, $8, $9, $10,
        $11, $12, $13,
        $14::jsonb, $15, $16, $17,
        $18, db_time.completed_at,
        CASE $8
            WHEN 'interactive_write' THEN db_time.completed_at + interval '24 hours'
            WHEN 'administrative_write' THEN db_time.completed_at + interval '30 days'
            WHEN 'async_handle' THEN GREATEST(
                db_time.completed_at + interval '30 days', $19::timestamptz
            )
        END
    FROM db_time
    RETURNING completed_at, retain_until
)
SELECT completed_at, retain_until FROM inserted;
