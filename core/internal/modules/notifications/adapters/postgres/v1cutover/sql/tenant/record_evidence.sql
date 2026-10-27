INSERT INTO notifications_cutover_evidence (
    id, run_id, source_queue_depth, future_delivery_count,
    running_delivery_count, unresolved_provider_receipt_count,
    legacy_connection_count, source_evidence_digest, observed_at, expires_at
)
SELECT $1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8::bytea,
       $9::timestamptz, $9::timestamptz + interval '5 minutes'
WHERE $9::timestamptz <= pg_catalog.transaction_timestamp()
  AND $9::timestamptz > pg_catalog.transaction_timestamp() - interval '5 minutes'
  AND EXISTS (
      SELECT 1 FROM notifications_writer_authority
      WHERE singleton AND writer = 'v1' AND lifecycle_state = 'quiescing'
  )
  AND EXISTS (
      SELECT 1 FROM notifications_adoption_runs
      WHERE id = $2::uuid AND lifecycle_state = 'verified'
  )
ON CONFLICT (run_id, source_evidence_digest) DO UPDATE
SET source_queue_depth = EXCLUDED.source_queue_depth,
    future_delivery_count = EXCLUDED.future_delivery_count,
    running_delivery_count = EXCLUDED.running_delivery_count,
    unresolved_provider_receipt_count = EXCLUDED.unresolved_provider_receipt_count,
    legacy_connection_count = EXCLUDED.legacy_connection_count,
    observed_at = EXCLUDED.observed_at, expires_at = EXCLUDED.expires_at
RETURNING id, expires_at;
