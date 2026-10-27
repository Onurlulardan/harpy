WITH inserted AS (
    INSERT INTO documents_forms_cleanup_operations (
        operation_id, definition_id, reference_epoch, lifecycle_state,
        catalog_version, source_owner, source_kind, usage,
        requested_at, updated_at
    )
    VALUES (
        $1::uuid, $2::uuid, $3, 'REQUESTED',
        'documents.forms-cleanup.v1', 'FORMS', 'form_record', 'LIFECYCLE',
        $4::timestamptz, $4::timestamptz
    )
    ON CONFLICT (operation_id) DO NOTHING
    RETURNING *
)
SELECT operation_id, definition_id, reference_epoch, lifecycle_state,
       link_count, item_count, version_count, ciphertext_bytes,
       link_catalog_sha256, zero_blocker_sha256, receipt_sha256,
       requested_at, completed_at
FROM inserted
UNION ALL
SELECT operation_id, definition_id, reference_epoch, lifecycle_state,
       link_count, item_count, version_count, ciphertext_bytes,
       link_catalog_sha256, zero_blocker_sha256, receipt_sha256,
       requested_at, completed_at
FROM documents_forms_cleanup_operations
WHERE operation_id = $1::uuid AND NOT EXISTS (SELECT 1 FROM inserted);
