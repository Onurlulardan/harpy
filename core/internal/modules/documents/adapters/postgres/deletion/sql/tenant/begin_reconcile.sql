WITH fence AS MATERIALIZED (
    SELECT pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended(
            'documents:dr-hold:' || pg_catalog.current_schema(), 0
        )
    )
)
UPDATE documents_blob_deletions AS deletion
SET lifecycle_state = 'RECONCILING', lease_owner_id = NULL,
    lease_expires_at = NULL, updated_at = pg_catalog.statement_timestamp()
WHERE deletion.id = $1::uuid AND deletion.lifecycle_state = 'CLAIMED'
  AND deletion.lease_owner_id = $2::uuid AND deletion.lease_generation = $3
  AND EXISTS (SELECT 1 FROM fence)
  AND (deletion.target_kind = 'UPLOAD_ATTEMPT' OR NOT EXISTS (
      SELECT 1 FROM documents_dr_holds AS hold
      WHERE hold.lifecycle_state IN ('HOLDING','COPYING','VERIFIED')
  ))
RETURNING true;
