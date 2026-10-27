WITH prior AS MATERIALIZED (
    SELECT request_sha256, aggregate_id
    FROM documents_operations
    WHERE actor_id = $2::uuid
      AND operation_kind = 'documents.legal_hold.release'
      AND idempotency_key = $7
), changed AS (
    UPDATE documents_legal_holds h
    SET lifecycle_state = 'RELEASED', released_at = $5
    WHERE h.id = $1::uuid AND h.lifecycle_state = 'ACTIVE'
      AND h.source_owner = $3 AND h.source_case_id = $4
      AND NOT EXISTS (SELECT 1 FROM prior)
    RETURNING h.*
), written_operation AS (
    INSERT INTO documents_operations (
        operation_id, operation_kind, actor_id, idempotency_key,
        request_sha256, aggregate_kind, aggregate_id, aggregate_version,
        response_status, response_document, occurred_at
    )
    SELECT $6::uuid, 'documents.legal_hold.release', $2::uuid, $7, $8,
           'LEGAL_HOLD', id, 2, 200,
           pg_catalog.convert_to(pg_catalog.jsonb_build_object(
               'id', id::text, 'state', 'released')::text, 'UTF8'), $5
    FROM changed RETURNING operation_id, aggregate_id
), written_audit AS (
    INSERT INTO documents_audit_entries (
        audit_id, operation_id, actor_id, action, aggregate_kind,
        aggregate_id, aggregate_version, occurred_at
    )
    SELECT $9::uuid, operation_id, $2::uuid, 'documents.legal_hold.release',
           'LEGAL_HOLD', aggregate_id, 2, $5 FROM written_operation
), result AS (
    SELECT h.*, false AS replay FROM changed h
    UNION ALL
    SELECT h.*, true AS replay FROM prior p
    JOIN documents_legal_holds h ON h.id = p.aggregate_id
    WHERE p.request_sha256 = $8
)
SELECT id, item_id, source_owner, source_case_id, evidence_sha256,
       lifecycle_state, placed_at, released_at, replay
FROM result;
