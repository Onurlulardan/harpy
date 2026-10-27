WITH attempt AS (
    UPDATE documents_upload_attempts a
    SET lifecycle_state = 'EFFECT_RECORDED', object_key = $4,
        object_version = $5, object_generation = $6,
        plaintext_bytes = $7, ciphertext_bytes = $8,
        plaintext_sha256 = $9, ciphertext_sha256 = $10,
        detected_mime = $11, file_kind = pg_catalog.upper($12),
        key_id = $13, wrapped_dek = $14,
        scan_signature_revision = $15, scanned_at = $16,
        updated_at = pg_catalog.statement_timestamp()
    FROM documents_uploads u
    WHERE a.upload_id = $1::uuid AND a.attempt_generation = $3
      AND a.lifecycle_state = 'RECEIVING'
      AND u.id = a.upload_id AND u.actor_id = $2::uuid
      AND u.lifecycle_state = 'RECEIVING'
      AND u.attempt_generation = a.attempt_generation
    RETURNING a.upload_id, a.attempt_generation
), changed AS (
    UPDATE documents_uploads u
    SET lifecycle_state = 'SCANNING', lease_owner_id = NULL,
        lease_expires_at = NULL, updated_at = pg_catalog.statement_timestamp()
    FROM attempt a
    WHERE u.id = a.upload_id AND u.attempt_generation = a.attempt_generation
    RETURNING u.id
)
SELECT true FROM changed;
