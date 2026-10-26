WITH prior AS MATERIALIZED (
    SELECT request_sha256, aggregate_id, response_document
    FROM documents_operations
    WHERE actor_id = $3::uuid AND operation_kind = 'documents.item.move'
      AND idempotency_key = $6
), changed AS (
    UPDATE documents_items i
    SET folder_id = $9::uuid, row_version = i.row_version + 1, updated_at = $8
    WHERE i.id = $4::uuid AND i.row_version = $5
      AND NOT EXISTS (SELECT 1 FROM prior)
      AND i.lifecycle_state = 'ACTIVE' AND EXISTS (SELECT 1 FROM documents_folders f WHERE f.id = $9::uuid AND f.lifecycle_state = 'ACTIVE')
    RETURNING i.id, i.row_version
), receipt AS (
    SELECT documents_item_response_v1(id) AS response_document, id, row_version
    FROM changed
), written_operation AS (
    INSERT INTO documents_operations (
        operation_id, operation_kind, actor_id, idempotency_key,
        request_sha256, aggregate_kind, aggregate_id, aggregate_version,
        response_status, response_document, occurred_at
    )
    SELECT $1::uuid, 'documents.item.move', $3::uuid, $6, $7,
           'ITEM', id, row_version, 200, response_document, $8
    FROM receipt
    RETURNING operation_id
), written_audit AS (
    INSERT INTO documents_audit_entries (
        audit_id, operation_id, actor_id, action, aggregate_kind,
        aggregate_id, aggregate_version, occurred_at
    )
    SELECT $2::uuid, operation_id, $3::uuid, 'documents.item.move', 'ITEM',
           receipt.id, receipt.row_version, $8
    FROM written_operation CROSS JOIN receipt
)
SELECT response_document, false FROM receipt
UNION ALL
SELECT response_document, true FROM prior
WHERE request_sha256 = $7 AND aggregate_id = $4::uuid;

