SELECT u.id, u.document_id, u.candidate_version_id, u.billing_reservation_id,
       u.attempt_generation, u.expected_plaintext_bytes, u.ciphertext_upper_bound,
       u.lifecycle_state, u.expires_at, u.lease_expires_at,
       a.attempt_generation, a.lifecycle_state, a.object_key, a.object_version,
       a.object_generation, a.ciphertext_bytes, a.updated_at
FROM documents_uploads u
LEFT JOIN documents_upload_attempts a
  ON a.upload_id = u.id AND a.attempt_generation = $2
WHERE u.id = $1::uuid;
