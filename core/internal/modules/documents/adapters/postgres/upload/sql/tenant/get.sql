SELECT id, actor_id, document_id, candidate_version_id,
       expected_current_version_id, folder_id, billing_reservation_id,
       request_sha256, expected_content_sha256, expected_plaintext_bytes,
       ciphertext_upper_bound, original_name, declared_mime, description, tags,
       attempt_generation, lifecycle_state, cleanup_required,
       stable_error_code, ready_status, ready_response, expires_at,
       created_at, updated_at, cancel_requested_at
FROM documents_uploads
WHERE id = $1::uuid AND actor_id = $2::uuid;
