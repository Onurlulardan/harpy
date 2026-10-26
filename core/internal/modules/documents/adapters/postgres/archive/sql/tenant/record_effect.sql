UPDATE documents_archives
SET artifact_plaintext_bytes = $2, ciphertext_bytes = $3,
    artifact_sha256 = $4, ciphertext_sha256 = $5, cipher_format = 'FCDOC1',
    object_key = $6, object_version = $7, object_generation = $8,
    key_id = $9, wrapped_dek = $10, updated_at = $11
WHERE id = $1::uuid AND lifecycle_state = 'PENDING' AND entry_count > 0
  AND object_version IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM documents_restore_transient_dispositions AS disposition
      WHERE disposition.target_kind = 'ARCHIVE'
        AND disposition.target_id = documents_archives.id
  )
RETURNING id;
