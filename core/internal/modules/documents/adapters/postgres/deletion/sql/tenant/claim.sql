WITH fence AS MATERIALIZED (
    SELECT pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended(
            'documents:dr-hold:' || pg_catalog.current_schema(), 0
        )
    )
), changed AS (
    UPDATE documents_blob_deletions d
    SET lifecycle_state = 'CLAIMED', lease_owner_id = $2::uuid,
        lease_generation = d.lease_generation + 1, lease_expires_at = $4,
        updated_at = $3
    WHERE d.id = $1::uuid
      AND (d.lifecycle_state = 'PENDING' OR
           (d.lifecycle_state = 'CLAIMED' AND d.lease_expires_at <= $3))
      AND EXISTS (SELECT 1 FROM fence)
      AND (d.target_kind = 'UPLOAD_ATTEMPT' OR NOT EXISTS (
          SELECT 1 FROM documents_dr_holds AS hold
          WHERE hold.lifecycle_state IN ('HOLDING','COPYING','VERIFIED')
      ))
    RETURNING d.*
)
SELECT d.id, d.target_kind, d.target_id, d.object_key, d.object_version,
       d.object_generation, d.plaintext_bytes, d.ciphertext_bytes,
       d.lease_owner_id, d.lease_generation, d.lease_expires_at,
       u.billing_reservation_id,
       CASE
         WHEN d.target_kind <> 'UPLOAD_ATTEMPT' THEN 'decrease'
         WHEN u.lifecycle_state = 'RETRYABLE' THEN 'keep'
         ELSE 'release'
       END AS billing_action
FROM changed d
LEFT JOIN documents_uploads u
  ON d.target_kind = 'UPLOAD_ATTEMPT' AND u.id = d.target_id;
