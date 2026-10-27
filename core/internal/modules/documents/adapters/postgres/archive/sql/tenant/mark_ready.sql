UPDATE documents_archives
SET lifecycle_state = 'READY', ready_at = $11::timestamptz,
    expires_at = $11::timestamptz + interval '24 hours',
    updated_at = $11::timestamptz
WHERE id = $1::uuid AND lifecycle_state = 'PENDING' AND entry_count > 0
  AND artifact_plaintext_bytes = $2 AND ciphertext_bytes = $3
  AND artifact_sha256 = $4 AND ciphertext_sha256 = $5
  AND object_key = $6 AND object_version = $7 AND object_generation = $8
  AND key_id = $9 AND wrapped_dek = $10
  AND NOT EXISTS (
      SELECT 1 FROM documents_restore_transient_dispositions AS disposition
      WHERE disposition.target_kind = 'ARCHIVE'
        AND disposition.target_id = documents_archives.id
  )
RETURNING id;
