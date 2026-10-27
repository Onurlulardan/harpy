INSERT INTO notifications_retention_receipts (
    job_id, attempt_id, tenant_id, payload_reference, payload_sha256,
    successor_job_id, successor_run_at, successor_dedupe_key,
    content_redacted, attempts_purged, receipts_purged, effects_purged,
    deliveries_purged, messages_purged, hints_purged,
    platform_records_purged, has_remaining, checkpoint_revision,
    receipt_sha256, completed_at
) VALUES (
    $1::uuid, $2::uuid, $3::uuid, $4, pg_catalog.decode($5, 'hex'),
    $6::uuid, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
    $17, $18, $19, $20
)
RETURNING job_id::text;
