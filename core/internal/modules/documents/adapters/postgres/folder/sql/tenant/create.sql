WITH prior AS MATERIALIZED (
    SELECT request_sha256, aggregate_id, response_document
    FROM documents_operations
    WHERE actor_id = $3::uuid AND operation_kind = 'documents.folder.create'
      AND idempotency_key = $7
), created AS (
    INSERT INTO documents_folders (
        id, parent_id, owner_id, display_name, lifecycle_state,
        row_version, created_at, updated_at
    )
    SELECT $4::uuid, $5::uuid, $3::uuid, $6, 'ACTIVE', 1, $9, $9
    WHERE NOT EXISTS (SELECT 1 FROM prior)
      AND ($5::uuid IS NULL OR EXISTS (
          SELECT 1 FROM documents_folders
          WHERE id = $5::uuid AND lifecycle_state = 'ACTIVE'
      ))
    RETURNING *
), receipt AS (
    SELECT pg_catalog.convert_to(pg_catalog.jsonb_build_object(
        'id', id::text, 'parent_id', parent_id::text, 'owner_id', owner_id::text,
        'display_name', display_name, 'state', lifecycle_state,
        'row_version', row_version, 'created_at', created_at,
        'updated_at', updated_at, 'trashed_at', trashed_at
    )::text, 'UTF8') AS response_document, id, row_version
    FROM created
), written_operation AS (
    INSERT INTO documents_operations (
        operation_id, operation_kind, actor_id, idempotency_key, request_sha256,
        aggregate_kind, aggregate_id, aggregate_version, response_status,
        response_document, occurred_at
    )
    SELECT $1::uuid, 'documents.folder.create', $3::uuid, $7, $8,
           'FOLDER', id, row_version, 201, response_document, $9
    FROM receipt
    RETURNING operation_id
), written_audit AS (
    INSERT INTO documents_audit_entries (
        audit_id, operation_id, actor_id, action, aggregate_kind,
        aggregate_id, aggregate_version, occurred_at
    )
    SELECT $2::uuid, operation_id, $3::uuid, 'documents.folder.create',
           'FOLDER', receipt.id, receipt.row_version, $9
    FROM written_operation CROSS JOIN receipt
)
SELECT response_document, false FROM receipt
UNION ALL
SELECT response_document, true FROM prior
WHERE request_sha256 = $8 AND aggregate_id = $4::uuid;
