SELECT id, actor_id, authorization_revision, job_id, retry_of_archive_id,
       request_sha256, billing_reservation_id, lifecycle_state,
       entry_count, plaintext_bytes, artifact_plaintext_bytes,
       ciphertext_bytes, entries_sha256, artifact_sha256, ciphertext_sha256,
       object_key, object_version, object_generation, key_id, wrapped_dek,
       ready_at, expires_at, created_at, updated_at
FROM documents_archives WHERE id = $1::uuid;
