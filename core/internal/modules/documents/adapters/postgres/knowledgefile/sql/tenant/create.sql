WITH source AS MATERIALIZED (
    SELECT u.id, u.actor_id, u.document_id, u.candidate_version_id,
           u.expected_plaintext_bytes, u.expected_content_sha256,
           u.original_name, u.declared_mime, u.expires_at
    FROM documents_uploads u
    WHERE u.id = $1::uuid AND u.actor_id = $2::uuid
      AND u.document_id = $3::uuid AND u.candidate_version_id = $4::uuid
      AND u.expected_plaintext_bytes = $5
      AND u.expected_content_sha256 = $6
      AND u.original_name = $7 AND u.declared_mime = $8
      AND u.expires_at = $9 AND u.lifecycle_state = 'CREATED'
), inserted AS (
    INSERT INTO documents_knowledge_file_leases (
        lease_id, upload_id, actor_id, item_id, version_id,
        expected_plaintext_bytes, expected_plaintext_sha256,
        filename, declared_mime, lifecycle_state,
        expires_at, created_at, updated_at
    )
    SELECT id, id, actor_id, document_id, candidate_version_id,
           expected_plaintext_bytes, expected_content_sha256,
           original_name, declared_mime, 'PENDING', expires_at, $10, $10
    FROM source
    ON CONFLICT (lease_id) DO NOTHING
    RETURNING *
), result AS (
    SELECT * FROM inserted
    UNION ALL
    SELECT l.* FROM documents_knowledge_file_leases l
    WHERE l.lease_id = $1::uuid AND l.actor_id = $2::uuid
      AND l.item_id = $3::uuid AND l.version_id = $4::uuid
      AND l.expected_plaintext_bytes = $5
      AND l.expected_plaintext_sha256 = $6
      AND l.filename = $7 AND l.declared_mime = $8 AND l.expires_at = $9
      AND NOT EXISTS (SELECT 1 FROM inserted)
)
SELECT lease_id, upload_id, actor_id, item_id, version_id,
       reference_id, source_id, expected_plaintext_bytes,
       expected_plaintext_sha256, filename, declared_mime,
       detected_mime, scan_signature_revision, scanned_at,
       lifecycle_state, expires_at, adopted_at, released_at
FROM result;
