WITH prior AS MATERIALIZED (
    SELECT request_sha256, aggregate_id
    FROM documents_operations
    WHERE actor_id = $2::uuid AND operation_kind = 'documents.link.release'
      AND idempotency_key = $9
), changed AS (
    UPDATE documents_links l
    SET lifecycle_state = 'RELEASED', released_at = $7
    WHERE l.id = $1::uuid AND l.lifecycle_state = 'ACTIVE'
      AND l.source_owner = $3 AND l.source_kind = $4
      AND l.source_id = $5::uuid AND l.source_version_id = $6::uuid
      AND NOT EXISTS (SELECT 1 FROM prior)
    RETURNING l.*
), written_operation AS (
    INSERT INTO documents_operations (
        operation_id, operation_kind, actor_id, idempotency_key,
        request_sha256, aggregate_kind, aggregate_id, aggregate_version,
        response_status, response_document, occurred_at
    )
    SELECT $8::uuid, 'documents.link.release', $2::uuid, $9, $10,
           'LINK', id, 2, 200,
           pg_catalog.convert_to(pg_catalog.jsonb_build_object(
               'id', id::text, 'state', 'released')::text, 'UTF8'), $7
    FROM changed RETURNING operation_id, aggregate_id
), written_audit AS (
    INSERT INTO documents_audit_entries (
        audit_id, operation_id, actor_id, action, aggregate_kind,
        aggregate_id, aggregate_version, occurred_at
    )
    SELECT $11::uuid, operation_id, $2::uuid, 'documents.link.release',
           'LINK', aggregate_id, 2, $7 FROM written_operation
), result AS (
    SELECT l.*, false AS replay FROM changed l
    UNION ALL
    SELECT l.*, true AS replay FROM prior p
    JOIN documents_links l ON l.id = p.aggregate_id
    WHERE p.request_sha256 = $10
)
SELECT id, item_id, source_owner, source_kind, source_id, source_version_id,
       usage, reference_epoch, lifecycle_state, created_at, released_at, replay
FROM result;
