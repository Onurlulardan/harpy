SELECT u.id, u.actor_id, u.document_id, u.candidate_version_id,
       u.expected_current_version_id, u.folder_id, u.billing_reservation_id,
       u.request_sha256, u.expected_content_sha256, u.expected_plaintext_bytes,
       u.ciphertext_upper_bound, u.original_name, u.declared_mime,
       u.description, u.tags,
       u.attempt_generation, u.lifecycle_state, u.cleanup_required,
       u.stable_error_code, u.ready_status, u.ready_response, u.expires_at,
       u.created_at, u.updated_at, u.cancel_requested_at
FROM documents_operations o
JOIN documents_uploads u ON u.id = o.aggregate_id
WHERE o.actor_id = $1::uuid
  AND o.operation_kind = 'documents.upload.create'
  AND o.idempotency_key = $2
  AND o.request_sha256 = $3
  AND o.aggregate_kind = 'UPLOAD';
