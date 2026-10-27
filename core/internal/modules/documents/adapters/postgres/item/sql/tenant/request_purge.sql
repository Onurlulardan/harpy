WITH prior AS MATERIALIZED (
    SELECT request_sha256, aggregate_id, response_document
    FROM documents_operations
    WHERE actor_id = $3::uuid AND operation_kind = 'documents.item.purge'
      AND idempotency_key = $6
), candidate AS MATERIALIZED (
    SELECT i.* FROM documents_items i
    WHERE i.id = $4::uuid AND i.row_version = $5
      AND i.lifecycle_state IN ('ACTIVE','TRASHED')
      AND NOT EXISTS (SELECT 1 FROM prior)
    FOR UPDATE
), held AS (
    SELECT true AS blocked FROM documents_legal_holds h
    JOIN candidate i ON i.id = h.item_id
    WHERE h.lifecycle_state = 'ACTIVE' LIMIT 1
), changed AS (
    UPDATE documents_items i
    SET resume_state = i.lifecycle_state, lifecycle_state = 'PURGE_COOLING',
        purge_requested_at = $8, purge_starts_at = $8 + interval '24 hours',
        purge_reason = $9, reference_epoch = i.reference_epoch + 1,
        row_version = i.row_version + 1, updated_at = $8
    FROM candidate c
    WHERE i.id = c.id AND NOT EXISTS (SELECT 1 FROM held)
      AND $10::timestamptz <= $8
      AND $8 - $10::timestamptz <= interval '10 minutes'
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
    SELECT $1::uuid, 'documents.item.purge', $3::uuid, $6, $7,
           'ITEM', id, row_version, 202, response_document, $8
    FROM receipt RETURNING operation_id
), written_audit AS (
    INSERT INTO documents_audit_entries (
        audit_id, operation_id, actor_id, action, aggregate_kind,
        aggregate_id, aggregate_version, reason, occurred_at
    )
    SELECT $2::uuid, operation_id, $3::uuid, 'documents.item.purge',
           'ITEM', receipt.id, receipt.row_version, $9, $8
    FROM written_operation CROSS JOIN receipt
)
SELECT response_document, false, false FROM receipt
UNION ALL
SELECT response_document, true, false FROM prior
WHERE request_sha256 = $7 AND aggregate_id = $4::uuid
UNION ALL
SELECT NULL::bytea, false, true FROM held;
