WITH prior AS MATERIALIZED (
    SELECT request_sha256, aggregate_id
    FROM documents_operations
    WHERE actor_id = $2::uuid AND operation_kind = 'documents.link.register'
      AND idempotency_key = $13
), target AS MATERIALIZED (
    SELECT id FROM documents_items
    WHERE id = $3::uuid AND lifecycle_state = 'ACTIVE' AND reference_epoch = $10
      AND NOT EXISTS (SELECT 1 FROM prior)
    FOR UPDATE
), inserted AS (
    INSERT INTO documents_links (
        id, item_id, source_owner, source_kind, source_id, source_version_id,
        usage, reference_epoch, lifecycle_state, created_at
    )
    SELECT $1::uuid, id, $4, $5, $6::uuid, $7::uuid,
           $8, $10, 'ACTIVE', $11 FROM target
    ON CONFLICT (item_id, source_owner, source_kind, source_id,
                 source_version_id, usage)
    WHERE lifecycle_state = 'ACTIVE'
    DO NOTHING
    RETURNING *
), written_operation AS (
    INSERT INTO documents_operations (
        operation_id, operation_kind, actor_id, idempotency_key,
        request_sha256, aggregate_kind, aggregate_id, aggregate_version,
        response_status, response_document, occurred_at
    )
    SELECT $9::uuid, 'documents.link.register', $2::uuid, $13, $12,
           'LINK', id, 1, 201,
           pg_catalog.convert_to(pg_catalog.jsonb_build_object(
               'id', id::text, 'state', 'active')::text, 'UTF8'), $11
    FROM inserted RETURNING operation_id, aggregate_id
), written_audit AS (
    INSERT INTO documents_audit_entries (
        audit_id, operation_id, actor_id, action, aggregate_kind,
        aggregate_id, aggregate_version, occurred_at
    )
    SELECT $14::uuid, operation_id, $2::uuid, 'documents.link.register',
           'LINK', aggregate_id, 1, $11 FROM written_operation
), result AS (
    SELECT l.*, false AS replay FROM inserted l
    UNION ALL
    SELECT l.*, true AS replay FROM prior p
    JOIN documents_links l ON l.id = p.aggregate_id
    WHERE p.request_sha256 = $12
)
SELECT id, item_id, source_owner, source_kind, source_id, source_version_id,
       usage, reference_epoch, lifecycle_state, created_at, released_at, replay
FROM result;
