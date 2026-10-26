SELECT attempt_id::text, tenant_id::text, payload_reference,
       pg_catalog.encode(payload_sha256, 'hex'), successor_job_id::text,
       successor_run_at, successor_dedupe_key, content_redacted,
       attempts_purged, receipts_purged, effects_purged,
       deliveries_purged, messages_purged, hints_purged,
       platform_records_purged, has_remaining, checkpoint_revision,
       receipt_sha256, completed_at
FROM notifications_retention_receipts
WHERE job_id = $1::uuid;
