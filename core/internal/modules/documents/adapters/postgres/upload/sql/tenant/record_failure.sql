WITH candidate AS MATERIALIZED (
    SELECT u.id, u.attempt_generation, u.cancel_requested_at, u.lifecycle_state
    FROM documents_uploads u
    WHERE u.id = $1::uuid AND u.actor_id = $2::uuid
      AND u.attempt_generation = $3
      AND u.lifecycle_state IN ('RECEIVING','SCANNING')
    FOR UPDATE
), deletion AS (
    INSERT INTO documents_blob_deletions (
        id, target_kind, target_id, object_key, object_version,
        object_generation, plaintext_bytes, ciphertext_bytes,
        lifecycle_state, lease_generation, created_at, updated_at
    )
    SELECT $7::uuid, 'UPLOAD_ATTEMPT', id, $8, $9, $10, $11, $12,
           'PENDING', 0, $6, $6
    FROM candidate WHERE $7::uuid IS NOT NULL
    ON CONFLICT (target_kind, target_id, object_generation) DO NOTHING
    RETURNING id
), changed_attempt AS (
    UPDATE documents_upload_attempts a
    SET lifecycle_state = CASE WHEN $7::uuid IS NULL
            THEN 'FAILED' ELSE 'CLEANUP_PENDING' END,
        object_key = $8, object_version = $9, object_generation = $10,
        plaintext_bytes = $11, ciphertext_bytes = $12,
        plaintext_sha256 = $13, ciphertext_sha256 = $14,
        key_id = $15, wrapped_dek = $16,
        cleanup_deletion_id = $7::uuid, updated_at = $6
    FROM candidate c
    WHERE a.upload_id = c.id AND a.attempt_generation = c.attempt_generation
      AND a.lifecycle_state IN ('RECEIVING','EFFECT_RECORDED')
      AND ($7::uuid IS NULL OR EXISTS (SELECT 1 FROM deletion))
    RETURNING a.upload_id
), changed_upload AS (
    UPDATE documents_uploads u
    SET lifecycle_state = CASE
            WHEN u.cancel_requested_at IS NOT NULL OR $5 = 'DOCUMENT_UPLOAD_CANCELLED'
                THEN 'CANCELLED'
            WHEN $5 = 'DOCUMENT_VERSION_CONFLICT' THEN 'CONFLICTED'
            WHEN $4 = 'transient' THEN 'RETRYABLE'
            ELSE 'REJECTED' END,
        cleanup_required = ($7::uuid IS NOT NULL), stable_error_code = $5,
        lease_owner_id = NULL, lease_expires_at = NULL, updated_at = $6
    FROM changed_attempt a
    WHERE u.id = a.upload_id
    RETURNING u.id
)
SELECT true FROM changed_upload;
