WITH prior AS MATERIALIZED (
    SELECT request_sha256, aggregate_id, response_document
    FROM documents_operations
    WHERE actor_id = $3::uuid AND operation_kind = 'documents.folder.delete'
      AND idempotency_key = $6
), changed AS (
    UPDATE documents_folders
    SET lifecycle_state = 'TRASHED', trashed_at = $8,
        row_version = row_version + 1, updated_at = $8
    WHERE id = $4::uuid AND row_version = $5 AND lifecycle_state = 'ACTIVE'
      AND NOT EXISTS (SELECT 1 FROM prior)
      AND NOT EXISTS (SELECT 1 FROM documents_folders child
          WHERE child.parent_id = $4::uuid AND child.lifecycle_state = 'ACTIVE')
      AND NOT EXISTS (SELECT 1 FROM documents_items item
          WHERE item.folder_id = $4::uuid AND item.lifecycle_state <> 'PURGED')
    RETURNING *
), receipt AS (
    SELECT pg_catalog.convert_to(pg_catalog.jsonb_build_object(
        'id', id::text, 'parent_id', parent_id::text, 'owner_id', owner_id::text,
        'display_name', display_name, 'state', lifecycle_state,
        'row_version', row_version, 'created_at', created_at,
        'updated_at', updated_at, 'trashed_at', trashed_at
    )::text, 'UTF8') AS response_document, id, row_version FROM changed
), written_operation AS (
    INSERT INTO documents_operations (
        operation_id, operation_kind, actor_id, idempotency_key, request_sha256,
        aggregate_kind, aggregate_id, aggregate_version, response_status,
        response_document, occurred_at
    ) SELECT $1::uuid, 'documents.folder.delete', $3::uuid, $6, $7,
        'FOLDER', id, row_version, 200, response_document, $8 FROM receipt
    RETURNING operation_id
), written_audit AS (
    INSERT INTO documents_audit_entries (
        audit_id, operation_id, actor_id, action, aggregate_kind,
        aggregate_id, aggregate_version, occurred_at
    ) SELECT $2::uuid, operation_id, $3::uuid, 'documents.folder.delete',
        'FOLDER', receipt.id, receipt.row_version, $8
      FROM written_operation CROSS JOIN receipt
)
SELECT response_document, false FROM receipt
UNION ALL SELECT response_document, true FROM prior
WHERE request_sha256 = $7 AND aggregate_id = $4::uuid;
