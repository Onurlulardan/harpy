WITH operation AS MATERIALIZED (
    SELECT * FROM documents_forms_cleanup_operations
    WHERE operation_id = $1::uuid FOR UPDATE
), version_evidence AS (
    SELECT count(version.id)::bigint AS version_count,
           COALESCE(pg_catalog.sum(version.ciphertext_bytes), 0)::bigint
               AS ciphertext_bytes,
           pg_catalog.sha256(pg_catalog.convert_to(
               COALESCE(pg_catalog.string_agg(
                   version.id::text || '|' ||
                   pg_catalog.encode(version.plaintext_sha256, 'hex') || '|' ||
                   pg_catalog.encode(version.delete_receipt_sha256, 'hex'),
                   '|' ORDER BY version.id
               ), ''), 'UTF8'
           )) AS version_catalog_sha256
    FROM documents_forms_cleanup_targets AS target
    JOIN documents_versions AS version ON version.item_id = target.item_id
    WHERE target.operation_id = $1::uuid
), readiness AS (
    SELECT NOT EXISTS (
               SELECT 1 FROM documents_links AS link, operation
               WHERE link.source_owner = 'FORMS'
                 AND link.source_kind = 'form_record'
                 AND link.source_id = operation.definition_id
                 AND link.usage = 'LIFECYCLE'
                 AND link.lifecycle_state = 'ACTIVE'
           )
           AND NOT EXISTS (
               SELECT 1 FROM documents_forms_cleanup_targets AS target
               JOIN documents_items AS item ON item.id = target.item_id
               WHERE target.operation_id = $1::uuid
                 AND item.lifecycle_state <> 'PURGED'
           )
           AND NOT EXISTS (
               SELECT 1 FROM documents_forms_cleanup_targets AS target
               JOIN documents_versions AS version ON version.item_id = target.item_id
               WHERE target.operation_id = $1::uuid
                 AND version.lifecycle_state <> 'PURGED'
           )
           AND NOT EXISTS (
               SELECT 1 FROM documents_forms_cleanup_targets AS target
               JOIN documents_versions AS version ON version.item_id = target.item_id
               WHERE target.operation_id = $1::uuid
                 AND NOT EXISTS (
                     SELECT 1 FROM documents_blob_deletions AS deletion
                     WHERE deletion.target_kind = 'VERSION'
                       AND deletion.target_id = version.id
                       AND deletion.lifecycle_state = 'CONFIRMED'
                 )
           ) AS ready
), evidence AS (
    SELECT operation.*,
           version_evidence.version_count AS computed_version_count,
           version_evidence.ciphertext_bytes AS computed_ciphertext_bytes,
           pg_catalog.sha256(pg_catalog.convert_to(
               operation.operation_id::text || '|' ||
               operation.definition_id::text || '|' ||
               operation.reference_epoch::text || '|' ||
               operation.link_count::text || '|' ||
               operation.item_count::text || '|' ||
               version_evidence.version_count::text || '|' ||
               version_evidence.ciphertext_bytes::text || '|' ||
               pg_catalog.encode(operation.link_catalog_sha256, 'hex') || '|' ||
               pg_catalog.encode(version_evidence.version_catalog_sha256, 'hex'),
               'UTF8'
           )) AS computed_zero_sha256
    FROM operation CROSS JOIN version_evidence CROSS JOIN readiness
    WHERE operation.lifecycle_state = 'RUNNING' AND readiness.ready
), receipted AS (
    SELECT evidence.*,
           pg_catalog.sha256(pg_catalog.convert_to(
               evidence.operation_id::text || '|' ||
               evidence.definition_id::text || '|' ||
               evidence.reference_epoch::text || '|' ||
               'documents.forms-cleanup.v1' || '|' ||
               pg_catalog.encode(evidence.computed_zero_sha256, 'hex') || '|' ||
               $2::timestamptz::text,
               'UTF8'
           )) AS final_receipt_sha256
    FROM evidence
), changed AS (
    UPDATE documents_forms_cleanup_operations AS cleanup
    SET lifecycle_state = 'SUCCEEDED',
        version_count = receipted.computed_version_count,
        ciphertext_bytes = receipted.computed_ciphertext_bytes,
        zero_blocker_sha256 = receipted.computed_zero_sha256,
        receipt_sha256 = receipted.final_receipt_sha256,
        completed_at = $2::timestamptz,
        updated_at = $2::timestamptz
    FROM receipted
    WHERE cleanup.operation_id = receipted.operation_id
    RETURNING cleanup.*
)
SELECT operation_id, definition_id, reference_epoch, lifecycle_state,
       link_count, item_count, version_count, ciphertext_bytes,
       link_catalog_sha256, zero_blocker_sha256, receipt_sha256,
       requested_at, completed_at
FROM changed
UNION ALL
SELECT operation_id, definition_id, reference_epoch, lifecycle_state,
       link_count, item_count, version_count, ciphertext_bytes,
       link_catalog_sha256, zero_blocker_sha256, receipt_sha256,
       requested_at, completed_at
FROM documents_forms_cleanup_operations
WHERE operation_id = $1::uuid AND NOT EXISTS (SELECT 1 FROM changed);
