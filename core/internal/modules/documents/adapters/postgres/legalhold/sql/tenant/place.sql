WITH prior AS MATERIALIZED (
    SELECT request_sha256, aggregate_id
    FROM documents_operations
    WHERE actor_id = $2::uuid
      AND operation_kind = 'documents.legal_hold.place'
      AND idempotency_key = $9
), target AS MATERIALIZED (
    SELECT id FROM documents_items
    WHERE id = $3::uuid AND lifecycle_state = 'ACTIVE'
      AND NOT EXISTS (SELECT 1 FROM prior)
    FOR UPDATE
), inserted AS (
    INSERT INTO documents_legal_holds (
        id, item_id, source_owner, source_case_id, evidence_sha256,
        lifecycle_state, placed_at
    )
    SELECT $1::uuid, id, $4, $5, $6, 'ACTIVE', $7 FROM target
    ON CONFLICT (item_id, source_owner, source_case_id) DO NOTHING
    RETURNING *
), written_operation AS (
    INSERT INTO documents_operations (
        operation_id, operation_kind, actor_id, idempotency_key,
        request_sha256, aggregate_kind, aggregate_id, aggregate_version,
        response_status, response_document, occurred_at
    )
    SELECT $8::uuid, 'documents.legal_hold.place', $2::uuid, $9, $10,
           'LEGAL_HOLD', id, 1, 201,
           pg_catalog.convert_to(pg_catalog.jsonb_build_object(
               'id', id::text, 'state', 'active')::text, 'UTF8'), $7
    FROM inserted RETURNING operation_id, aggregate_id
), written_audit AS (
    INSERT INTO documents_audit_entries (
        audit_id, operation_id, actor_id, action, aggregate_kind,
        aggregate_id, aggregate_version, occurred_at
    )
    SELECT $11::uuid, operation_id, $2::uuid, 'documents.legal_hold.place',
           'LEGAL_HOLD', aggregate_id, 1, $7 FROM written_operation
), result AS (
    SELECT h.*, false AS replay FROM inserted h
    UNION ALL
    SELECT h.*, true AS replay FROM prior p
    JOIN documents_legal_holds h ON h.id = p.aggregate_id
    WHERE p.request_sha256 = $10
)
SELECT id, item_id, source_owner, source_case_id, evidence_sha256,
       lifecycle_state, placed_at, released_at, replay
FROM result;
