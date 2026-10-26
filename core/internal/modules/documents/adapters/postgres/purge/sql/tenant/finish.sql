WITH candidate AS MATERIALIZED (
    SELECT i.* FROM documents_items i
    WHERE i.id = $1::uuid AND i.lifecycle_state = 'PURGING'
      AND NOT EXISTS (SELECT 1 FROM documents_legal_holds h
          WHERE h.item_id = i.id AND h.lifecycle_state = 'ACTIVE')
      AND NOT EXISTS (SELECT 1 FROM documents_versions v
          WHERE v.item_id = i.id AND v.lifecycle_state <> 'PURGED')
    FOR UPDATE
), evidence AS (
    SELECT pg_catalog.sha256(pg_catalog.convert_to(pg_catalog.string_agg(
        v.id::text || ':' || pg_catalog.encode(v.plaintext_sha256, 'hex') || ':' ||
        pg_catalog.encode(v.delete_receipt_sha256, 'hex'), '|' ORDER BY v.version_number
    ), 'UTF8')) AS digest
    FROM documents_versions v JOIN candidate i ON i.id = v.item_id
    HAVING count(*) > 0
), changed AS (
    UPDATE documents_items i
    SET lifecycle_state = 'PURGED', resume_state = NULL,
        trashed_at = NULL, retention_ends_at = NULL, purged_at = $2::timestamptz,
        tombstone_until = $2::timestamptz + interval '365 days',
        purge_evidence_sha256 = evidence.digest,
        row_version = i.row_version + 1, updated_at = $2::timestamptz
    FROM candidate c CROSS JOIN evidence
    WHERE i.id = c.id
    RETURNING i.id, i.owner_id, i.row_version, i.purged_at
), written_operation AS (
    INSERT INTO documents_operations (
        operation_id, operation_kind, actor_id, idempotency_key,
        request_sha256, aggregate_kind, aggregate_id, aggregate_version,
        response_status, response_document, occurred_at
    )
    SELECT $3::uuid, 'documents.item.purge.finish', owner_id,
           'purge-finish:' || id::text,
           pg_catalog.sha256(pg_catalog.convert_to(id::text, 'UTF8')),
           'ITEM', id, row_version, 200,
           pg_catalog.convert_to(pg_catalog.jsonb_build_object(
               'id', id::text, 'state', 'purged')::text, 'UTF8'), $2::timestamptz
    FROM changed
    RETURNING operation_id, actor_id, aggregate_id, aggregate_version
), written_audit AS (
    INSERT INTO documents_audit_entries (
        audit_id, operation_id, actor_id, action, aggregate_kind,
        aggregate_id, aggregate_version, occurred_at
    )
    SELECT $4::uuid, operation_id, actor_id, 'documents.item.purge.finish',
           'ITEM', aggregate_id, aggregate_version, $2::timestamptz FROM written_operation
)
SELECT changed.id, changed.purged_at FROM changed
JOIN written_operation ON written_operation.aggregate_id = changed.id;
