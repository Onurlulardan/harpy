UPDATE notifications_writer_authority authority
SET lifecycle_state = 'fenced', row_version = row_version + 1,
    updated_at = pg_catalog.transaction_timestamp()
WHERE authority.singleton AND authority.writer = 'v1'
  AND authority.lifecycle_state = 'quiescing'
  AND authority.generation = $1 AND authority.row_version = $2
  AND EXISTS (
      SELECT 1
      FROM notifications_cutover_evidence evidence
      JOIN notifications_adoption_runs run ON run.id = evidence.run_id
      WHERE evidence.id = $3::uuid AND evidence.run_id = $4::uuid
        AND evidence.source_evidence_digest = $5::bytea
        AND evidence.expires_at > pg_catalog.transaction_timestamp()
        AND evidence.source_queue_depth = 0
        AND evidence.future_delivery_count = 0
        AND evidence.running_delivery_count = 0
        AND evidence.unresolved_provider_receipt_count = 0
        AND evidence.legacy_connection_count = 0
        AND run.lifecycle_state = 'verified'
        AND NOT EXISTS (
            SELECT 1 FROM notifications_adoption_issues issue
            WHERE issue.run_id = run.id AND issue.resolved_at IS NULL
        )
  )
RETURNING writer, lifecycle_state, generation, row_version;
