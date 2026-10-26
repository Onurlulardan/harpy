WITH prior AS MATERIALIZED (
    SELECT request_sha256, aggregate_id
    FROM documents_operations
    WHERE actor_id = $2::uuid AND operation_kind = 'documents.archive.create'
      AND idempotency_key = $11
), inserted AS (
    INSERT INTO documents_archives (
        id, actor_id, authorization_revision, job_id, retry_of_archive_id,
        request_sha256, request_document, billing_reservation_id,
        lifecycle_state, entry_count, plaintext_bytes, created_at, updated_at
    )
    SELECT $1::uuid, $2::uuid, $3, $4::uuid, $5::uuid, $6, $7,
           $8::uuid, 'PENDING', 0, 0, $9, $9
    WHERE NOT EXISTS (SELECT 1 FROM prior)
      AND ($5::uuid IS NULL OR EXISTS (
          SELECT 1 FROM documents_archives WHERE id = $5::uuid
      ))
    ON CONFLICT (actor_id, request_sha256) DO NOTHING
    RETURNING id
), operation_receipt AS (
    SELECT pg_catalog.convert_to(pg_catalog.jsonb_build_object(
        'archive_id', id::text, 'state', 'pending'
    )::text, 'UTF8') AS response_document, id FROM inserted
), written_operation AS (
    INSERT INTO documents_operations (
        operation_id, operation_kind, actor_id, idempotency_key,
        request_sha256, aggregate_kind, aggregate_id, aggregate_version,
        response_status, response_document, occurred_at
    )
    SELECT $10::uuid, 'documents.archive.create', $2::uuid, $11, $6,
           'ARCHIVE', id, 1, 202, response_document, $9
    FROM operation_receipt RETURNING operation_id, aggregate_id
), written_audit AS (
    INSERT INTO documents_audit_entries (
        audit_id, operation_id, actor_id, action, aggregate_kind,
        aggregate_id, aggregate_version, occurred_at
    )
    SELECT $12::uuid, operation_id, $2::uuid, 'documents.archive.create',
           'ARCHIVE', aggregate_id, 1, $9 FROM written_operation
)
SELECT id FROM inserted;
