WITH prior AS MATERIALIZED (
    SELECT aggregate_id, request_sha256
    FROM documents_operations
    WHERE actor_id = $2::uuid
      AND operation_kind = 'documents.upload.create'
      AND idempotency_key = $20
), inserted AS (
    INSERT INTO documents_uploads (
        id, actor_id, document_id, candidate_version_id,
        expected_current_version_id, folder_id, request_sha256,
        expected_content_sha256, expected_plaintext_bytes,
        ciphertext_upper_bound, original_name, declared_mime, description, tags,
        billing_reservation_id, attempt_generation, lifecycle_state,
        cleanup_required, expires_at, created_at, updated_at
    )
    SELECT $1::uuid, $2::uuid, $3::uuid, $4::uuid, $5::uuid, $6::uuid,
           $7, $8, $9, $10, $11, $12, $13, $14::jsonb,
           $15::uuid, 0, 'CREATED', false, $16, $17, $17
    WHERE NOT EXISTS (SELECT 1 FROM prior)
      AND EXISTS (
          SELECT 1 FROM documents_folders
          WHERE id = $6::uuid AND lifecycle_state = 'ACTIVE'
      )
      AND (($5::uuid IS NULL AND NOT EXISTS (
              SELECT 1 FROM documents_items WHERE id = $3::uuid
           )) OR EXISTS (
              SELECT 1 FROM documents_items
              WHERE id = $3::uuid AND folder_id = $6::uuid
                AND lifecycle_state = 'ACTIVE'
                AND current_version_id = $5::uuid
           ))
    ON CONFLICT (actor_id, request_sha256) DO NOTHING
    RETURNING id
), operation_receipt AS (
    SELECT pg_catalog.convert_to(pg_catalog.jsonb_build_object(
        'upload_id', id::text, 'state', 'created'
    )::text, 'UTF8') AS response_document, id
    FROM inserted
), written_operation AS (
    INSERT INTO documents_operations (
        operation_id, operation_kind, actor_id, idempotency_key,
        request_sha256, aggregate_kind, aggregate_id, aggregate_version,
        response_status, response_document, occurred_at
    )
    SELECT $18::uuid, 'documents.upload.create', $2::uuid, $20, $7,
           'UPLOAD', id, 1, 201, response_document, $17
    FROM operation_receipt
    RETURNING operation_id, aggregate_id
), written_audit AS (
    INSERT INTO documents_audit_entries (
        audit_id, operation_id, actor_id, action, aggregate_kind,
        aggregate_id, aggregate_version, occurred_at
    )
    SELECT $19::uuid, operation_id, $2::uuid, 'documents.upload.create',
           'UPLOAD', aggregate_id, 1, $17
    FROM written_operation
)
SELECT u.id, u.actor_id, u.document_id, u.candidate_version_id,
       u.expected_current_version_id, u.folder_id, u.billing_reservation_id,
       u.request_sha256, u.expected_content_sha256, u.expected_plaintext_bytes,
       u.ciphertext_upper_bound, u.original_name, u.declared_mime,
       u.description, u.tags,
       u.attempt_generation, u.lifecycle_state, u.cleanup_required,
       u.stable_error_code, u.ready_status, u.ready_response, u.expires_at,
       u.created_at, u.updated_at, u.cancel_requested_at
FROM documents_uploads u
JOIN written_operation o ON o.aggregate_id = u.id;
