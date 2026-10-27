WITH candidate AS MATERIALIZED (
    SELECT u.*
    FROM documents_uploads u
    JOIN documents_upload_attempts a
      ON a.upload_id = u.id AND a.attempt_generation = u.attempt_generation
    WHERE u.id = $1::uuid AND u.actor_id = $2::uuid
      AND u.attempt_generation = $3 AND u.lifecycle_state = 'SCANNING'
      AND u.cancel_requested_at IS NULL
      AND a.lifecycle_state = 'EFFECT_RECORDED'
      AND a.object_key = $4 AND a.object_version = $5
      AND a.object_generation = $6 AND a.plaintext_bytes = $7
      AND a.ciphertext_bytes = $8 AND a.plaintext_sha256 = $9
      AND a.ciphertext_sha256 = $10
    FOR UPDATE OF u
), locked_existing AS MATERIALIZED (
    SELECT i.*
    FROM documents_items i JOIN candidate c ON c.document_id = i.id
    FOR UPDATE OF i
), seeded AS (
    INSERT INTO documents_items (
        id, folder_id, owner_id, current_version_id, lifecycle_state,
        row_version, description, tags, created_at, updated_at
    )
    SELECT c.document_id, c.folder_id, c.actor_id, c.candidate_version_id, 'ACTIVE', 1,
           c.description, c.tags,
           pg_catalog.statement_timestamp(), pg_catalog.statement_timestamp()
    FROM candidate c
    WHERE c.expected_current_version_id IS NULL
      AND NOT EXISTS (SELECT 1 FROM locked_existing)
    ON CONFLICT (id) DO NOTHING
    RETURNING id, folder_id, row_version
), eligible AS (
    SELECT id, folder_id, row_version FROM seeded
    UNION ALL
    SELECT i.id, i.folder_id, i.row_version
    FROM locked_existing i JOIN candidate c ON c.document_id = i.id
    WHERE i.lifecycle_state = 'ACTIVE' AND i.folder_id = c.folder_id
      AND i.current_version_id = c.expected_current_version_id
), inserted_version AS (
    INSERT INTO documents_versions (
        id, item_id, version_number, lifecycle_state, original_name,
        declared_mime, detected_mime, file_kind, plaintext_bytes,
        ciphertext_bytes, plaintext_sha256, ciphertext_sha256,
        cipher_format, object_key, object_version, object_generation,
        key_id, wrapped_dek, scan_signature_revision, scanned_at, created_at
    )
    SELECT c.candidate_version_id, c.document_id,
           (SELECT COALESCE(pg_catalog.max(v.version_number), 0) + 1
            FROM documents_versions v WHERE v.item_id = c.document_id),
           'READY', c.original_name, c.declared_mime, $11,
           pg_catalog.upper($12), $7, $8, $9, $10, 'FCDOC1', $4, $5,
           $6, $13, $14, $15, $16, pg_catalog.statement_timestamp()
    FROM candidate c JOIN eligible i ON i.id = c.document_id
    RETURNING id, item_id, version_number
), changed_existing AS (
    UPDATE documents_items i
    SET current_version_id = v.id, row_version = i.row_version + 1,
        updated_at = pg_catalog.statement_timestamp()
    FROM inserted_version v
    WHERE i.id = v.item_id AND i.current_version_id IS DISTINCT FROM v.id
    RETURNING i.id, i.folder_id, i.current_version_id, i.lifecycle_state,
              i.row_version, v.version_number
), item_result AS (
    SELECT s.id, s.folder_id, v.id AS current_version_id,
           'ACTIVE'::text AS lifecycle_state, s.row_version, v.version_number
    FROM seeded s JOIN inserted_version v ON v.item_id = s.id
    UNION ALL
    SELECT * FROM changed_existing
), response AS (
    SELECT pg_catalog.convert_to(pg_catalog.jsonb_build_object(
        'document_id', i.id::text, 'folder_id', i.folder_id::text,
        'current_version_id', i.current_version_id::text,
        'version_number', i.version_number, 'state', pg_catalog.lower(i.lifecycle_state),
        'row_version', i.row_version
    )::text, 'UTF8') AS body, i.*
    FROM item_result i
), changed_upload AS (
    UPDATE documents_uploads u
    SET lifecycle_state = 'READY', cleanup_required = false,
        stable_error_code = NULL, ready_status = 201, ready_response = r.body,
        lease_owner_id = NULL, lease_expires_at = NULL,
        updated_at = pg_catalog.statement_timestamp()
    FROM response r
    WHERE u.id = $1::uuid
    RETURNING u.id, u.actor_id, u.ready_status, u.ready_response,
              r.row_version, r.current_version_id
), changed_attempt AS (
    UPDATE documents_upload_attempts a
    SET lifecycle_state = 'FINALIZED', updated_at = pg_catalog.statement_timestamp()
    FROM changed_upload u
    WHERE a.upload_id = u.id AND a.attempt_generation = $3
      AND a.lifecycle_state = 'EFFECT_RECORDED'
    RETURNING a.upload_id
), written_operation AS (
    INSERT INTO documents_operations (
        operation_id, operation_kind, actor_id, idempotency_key,
        request_sha256, aggregate_kind, aggregate_id, aggregate_version,
        response_status, response_document, occurred_at
    )
    SELECT $17::uuid, 'documents.upload.put', u.actor_id, u.id::text,
           c.request_sha256, 'ITEM', c.document_id, u.row_version,
           201, u.ready_response, pg_catalog.statement_timestamp()
    FROM changed_upload u JOIN candidate c ON c.id = u.id
    JOIN changed_attempt a ON a.upload_id = u.id
    RETURNING operation_id, response_status, response_document, aggregate_id,
              aggregate_version, actor_id
), written_audit AS (
    INSERT INTO documents_audit_entries (
        audit_id, operation_id, actor_id, action, aggregate_kind,
        aggregate_id, aggregate_version, occurred_at
    )
    SELECT $18::uuid, operation_id, actor_id, 'documents.upload.ready',
           'ITEM', aggregate_id, aggregate_version, pg_catalog.statement_timestamp()
    FROM written_operation
)
SELECT response_status, response_document FROM written_operation;
