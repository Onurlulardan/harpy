WITH candidate AS MATERIALIZED (
    SELECT u.*, a.lifecycle_state AS attempt_state,
           a.object_version AS recorded_version
    FROM documents_uploads u
    JOIN documents_upload_attempts a
      ON a.upload_id = u.id AND a.attempt_generation = u.attempt_generation
    WHERE u.id = $1::uuid AND u.attempt_generation = $2
      AND u.lifecycle_state IN ('RECEIVING','SCANNING')
      AND a.lifecycle_state IN ('RECEIVING','EFFECT_RECORDED')
      AND (u.lifecycle_state = 'RECEIVING' AND u.lease_expires_at <= $8 OR
           u.lifecycle_state = 'SCANNING' AND a.updated_at <= $8 - interval '1 hour')
      AND ($4::text IS NULL OR a.object_version IS NULL OR a.object_version = $4)
    FOR UPDATE OF u, a
), deletion AS (
    INSERT INTO documents_blob_deletions (
        id, target_kind, target_id, object_key, object_version,
        object_generation, plaintext_bytes, ciphertext_bytes,
        lifecycle_state, lease_generation, created_at, updated_at
    )
    SELECT $7::uuid, 'UPLOAD_ATTEMPT', id, $3, $4, $2, $5, $6,
           'PENDING', 0, $8, $8
    FROM candidate WHERE $7::uuid IS NOT NULL
    ON CONFLICT (target_kind, target_id, object_generation) DO NOTHING
    RETURNING id
), changed_attempt AS (
    UPDATE documents_upload_attempts a
    SET lifecycle_state = CASE WHEN $4::text IS NULL
            THEN 'FAILED' ELSE 'CLEANUP_PENDING' END,
        object_key = CASE WHEN $4::text IS NULL THEN NULL ELSE $3 END,
        object_version = $4,
        object_generation = CASE WHEN $4::text IS NULL THEN NULL ELSE $2 END,
        plaintext_bytes = CASE WHEN $4::text IS NULL THEN NULL ELSE $5 END,
        ciphertext_bytes = CASE WHEN $4::text IS NULL THEN NULL ELSE $6 END,
        cleanup_deletion_id = $7::uuid, updated_at = $8
    FROM candidate c
    WHERE a.upload_id = c.id AND a.attempt_generation = c.attempt_generation
      AND ($7::uuid IS NULL OR EXISTS (SELECT 1 FROM deletion))
    RETURNING a.upload_id
), changed_upload AS (
    UPDATE documents_uploads u
    SET lifecycle_state = CASE
            WHEN u.cancel_requested_at IS NOT NULL THEN 'CANCELLED'
            WHEN u.expires_at <= $8 THEN 'EXPIRED'
            ELSE 'RETRYABLE' END,
        cleanup_required = ($7::uuid IS NOT NULL),
        stable_error_code = CASE WHEN $7::uuid IS NULL
            THEN 'DOCUMENT_ORPHAN_ABSENT' ELSE 'DOCUMENT_ORPHAN_CLEANUP_PENDING' END,
        lease_owner_id = NULL, lease_expires_at = NULL, updated_at = $8
    FROM changed_attempt a WHERE u.id = a.upload_id
    RETURNING u.id, u.billing_reservation_id, u.lifecycle_state,
              ($7::uuid IS NULL AND u.lifecycle_state IN ('CANCELLED','EXPIRED')) AS release
)
SELECT id, billing_reservation_id, lifecycle_state, release FROM changed_upload;
