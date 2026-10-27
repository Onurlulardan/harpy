WITH candidate AS MATERIALIZED (
    SELECT u.*
    FROM documents_uploads u
    WHERE u.id = $1::uuid AND u.actor_id = $2::uuid
      AND u.lifecycle_state IN ('CREATED','RETRYABLE')
      AND NOT u.cleanup_required AND u.cancel_requested_at IS NULL
      AND u.expires_at > pg_catalog.statement_timestamp()
      AND EXISTS (
          SELECT 1 FROM documents_folders f
          WHERE f.id = u.folder_id AND f.lifecycle_state = 'ACTIVE'
      )
      AND ((u.expected_current_version_id IS NULL AND NOT EXISTS (
              SELECT 1 FROM documents_items i WHERE i.id = u.document_id
           )) OR EXISTS (
              SELECT 1 FROM documents_items i
              WHERE i.id = u.document_id AND i.folder_id = u.folder_id
                AND i.lifecycle_state = 'ACTIVE'
                AND i.current_version_id = u.expected_current_version_id
           ))
    FOR UPDATE
), changed AS (
    UPDATE documents_uploads u
    SET lifecycle_state = 'RECEIVING',
        attempt_generation = u.attempt_generation + 1,
        lease_owner_id = $3::uuid,
        lease_expires_at = pg_catalog.statement_timestamp() + interval '5 minutes',
        updated_at = pg_catalog.statement_timestamp()
    FROM candidate c
    WHERE u.id = c.id
    RETURNING u.*
), attempt AS (
    INSERT INTO documents_upload_attempts (
        upload_id, attempt_generation, effect_receipt_id,
        lifecycle_state, created_at, updated_at
    )
    SELECT id, attempt_generation, $4::uuid, 'RECEIVING',
           pg_catalog.statement_timestamp(), pg_catalog.statement_timestamp()
    FROM changed
    RETURNING upload_id
)
SELECT c.id, c.actor_id, c.document_id, c.candidate_version_id,
       c.expected_current_version_id, c.billing_reservation_id,
       c.attempt_generation, c.original_name, c.declared_mime,
       c.expected_content_sha256, c.expected_plaintext_bytes
FROM changed c
JOIN attempt a ON a.upload_id = c.id;
