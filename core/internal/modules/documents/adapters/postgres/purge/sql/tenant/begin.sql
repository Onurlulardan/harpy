WITH tasks AS MATERIALIZED (
    SELECT deletion_id, version_id
    FROM pg_catalog.jsonb_to_recordset($3::jsonb)
      AS task(deletion_id uuid, version_id uuid)
), candidate AS MATERIALIZED (
    SELECT i.* FROM documents_items i
    WHERE i.id = $1::uuid
      AND ((i.lifecycle_state = 'PURGE_COOLING' AND i.purge_starts_at <= $2)
        OR (i.lifecycle_state = 'TRASHED' AND i.retention_ends_at <= $2))
      AND NOT EXISTS (SELECT 1 FROM documents_legal_holds h
          WHERE h.item_id = i.id AND h.lifecycle_state = 'ACTIVE')
      AND NOT EXISTS (SELECT 1 FROM documents_links l
          WHERE l.item_id = i.id AND l.lifecycle_state = 'ACTIVE'
            AND l.usage = 'LIFECYCLE')
      AND (SELECT count(*) FROM documents_versions v
           WHERE v.item_id = i.id AND v.lifecycle_state = 'READY') =
          (SELECT count(*) FROM tasks)
      AND NOT EXISTS (SELECT 1 FROM tasks t
          LEFT JOIN documents_versions v ON v.id = t.version_id
            AND v.item_id = i.id AND v.lifecycle_state = 'READY'
          WHERE v.id IS NULL)
    FOR UPDATE
), changed_item AS (
    UPDATE documents_items i
    SET lifecycle_state = 'PURGING', current_version_id = NULL,
        reference_epoch = CASE WHEN i.lifecycle_state = 'TRASHED'
            THEN i.reference_epoch + 1 ELSE i.reference_epoch END,
        row_version = i.row_version + 1, updated_at = $2
    FROM candidate c WHERE i.id = c.id
    RETURNING i.id, i.owner_id, i.row_version
), changed_versions AS (
    UPDATE documents_versions v SET lifecycle_state = 'DELETE_PENDING'
    FROM tasks t, changed_item i
    WHERE v.id = t.version_id AND v.item_id = i.id AND v.lifecycle_state = 'READY'
    RETURNING v.*
), inserted_deletions AS (
    INSERT INTO documents_blob_deletions (
        id, target_kind, target_id, object_key, object_version,
        object_generation, plaintext_bytes, ciphertext_bytes,
        lifecycle_state, lease_generation, created_at, updated_at
    )
    SELECT t.deletion_id, 'VERSION', v.id, v.object_key, v.object_version,
           v.object_generation, v.plaintext_bytes, v.ciphertext_bytes,
           'PENDING', 0, $2, $2
    FROM changed_versions v JOIN tasks t ON t.version_id = v.id
    RETURNING target_id
), written_operation AS (
    INSERT INTO documents_operations (
        operation_id, operation_kind, actor_id, idempotency_key,
        request_sha256, aggregate_kind, aggregate_id, aggregate_version,
        response_status, response_document, occurred_at
    )
    SELECT $4::uuid, 'documents.item.purge.begin', owner_id,
           'purge-begin:' || id::text,
           pg_catalog.sha256(pg_catalog.convert_to($3::jsonb::text, 'UTF8')),
           'ITEM', id, row_version, 202,
           pg_catalog.convert_to(pg_catalog.jsonb_build_object(
               'id', id::text, 'state', 'purging')::text, 'UTF8'), $2
    FROM changed_item
    WHERE (SELECT count(*) FROM inserted_deletions) = (SELECT count(*) FROM tasks)
    RETURNING operation_id, actor_id, aggregate_id, aggregate_version
), written_audit AS (
    INSERT INTO documents_audit_entries (
        audit_id, operation_id, actor_id, action, aggregate_kind,
        aggregate_id, aggregate_version, occurred_at
    )
    SELECT $5::uuid, operation_id, actor_id, 'documents.item.purge.begin',
           'ITEM', aggregate_id, aggregate_version, $2 FROM written_operation
)
SELECT aggregate_id, (SELECT count(*) FROM inserted_deletions)
FROM written_operation;
