WITH changed AS (
    UPDATE documents_archives
    SET lifecycle_state = 'EXPIRED', updated_at = $3
    WHERE id = $1::uuid AND lifecycle_state = 'READY' AND expires_at <= $3
    RETURNING *
), deletion AS (
    INSERT INTO documents_blob_deletions (
        id, target_kind, target_id, object_key, object_version,
        object_generation, plaintext_bytes, ciphertext_bytes,
        lifecycle_state, lease_generation, created_at, updated_at
    )
    SELECT $2::uuid, 'ARCHIVE', id, object_key, object_version,
           object_generation, artifact_plaintext_bytes, ciphertext_bytes,
           'PENDING', 0, $3, $3 FROM changed
    ON CONFLICT (target_kind, target_id, object_generation) DO NOTHING
    RETURNING target_id
)
SELECT id FROM changed JOIN deletion ON deletion.target_id = changed.id;
