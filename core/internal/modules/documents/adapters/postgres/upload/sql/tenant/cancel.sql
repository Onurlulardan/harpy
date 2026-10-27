WITH candidate AS MATERIALIZED (
    SELECT * FROM documents_uploads
    WHERE id = $1::uuid AND actor_id = $2::uuid
    FOR UPDATE
), changed AS (
    UPDATE documents_uploads u
    SET lifecycle_state = CASE
            WHEN u.lifecycle_state IN ('CREATED','RETRYABLE') THEN 'CANCELLED'
            ELSE u.lifecycle_state END,
        cancel_requested_at = CASE
            WHEN u.lifecycle_state IN ('RECEIVING','SCANNING')
                THEN COALESCE(u.cancel_requested_at, $4)
            ELSE u.cancel_requested_at END,
        updated_at = $4
    FROM candidate c
    WHERE u.id = c.id
      AND ((u.lifecycle_state IN ('CREATED','RETRYABLE') AND NOT u.cleanup_required)
           OR u.lifecycle_state IN ('RECEIVING','SCANNING'))
    RETURNING u.*,
      (c.lifecycle_state IN ('CREATED','RETRYABLE')) AS release_reservation
), result AS (
    SELECT * FROM changed
    UNION ALL
    SELECT c.*, false AS release_reservation FROM candidate c
    WHERE NOT EXISTS (SELECT 1 FROM changed)
)
SELECT id, actor_id, document_id, candidate_version_id,
       expected_current_version_id, folder_id, billing_reservation_id,
       request_sha256, expected_content_sha256, expected_plaintext_bytes,
       ciphertext_upper_bound, original_name, declared_mime, description, tags,
       attempt_generation, lifecycle_state, cleanup_required,
       stable_error_code, ready_status, ready_response, expires_at,
       created_at, updated_at, cancel_requested_at, release_reservation
FROM result;
