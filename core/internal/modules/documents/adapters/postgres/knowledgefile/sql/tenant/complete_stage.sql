WITH candidate AS MATERIALIZED (
    SELECT l.lease_id, v.detected_mime, v.scan_signature_revision, v.scanned_at
    FROM documents_knowledge_file_leases l
    JOIN documents_uploads u ON u.id = l.upload_id
    JOIN documents_items i ON i.id = l.item_id
    JOIN documents_versions v ON v.id = l.version_id AND v.item_id = i.id
    WHERE l.lease_id = $1::uuid AND l.actor_id = $2::uuid
      AND l.lifecycle_state IN ('PENDING','STAGED') AND l.expires_at > $3
      AND u.lifecycle_state = 'READY' AND i.lifecycle_state = 'ACTIVE'
      AND i.current_version_id = l.version_id AND v.lifecycle_state = 'READY'
      AND v.plaintext_bytes = l.expected_plaintext_bytes
      AND v.plaintext_sha256 = l.expected_plaintext_sha256
      AND v.original_name = l.filename AND v.declared_mime = l.declared_mime
    FOR UPDATE OF l
), changed AS (
    UPDATE documents_knowledge_file_leases l
    SET detected_mime = c.detected_mime,
        scan_signature_revision = c.scan_signature_revision,
        scanned_at = c.scanned_at, lifecycle_state = 'STAGED', updated_at = $3
    FROM candidate c
    WHERE l.lease_id = c.lease_id AND l.lifecycle_state = 'PENDING'
    RETURNING l.*
), result AS (
    SELECT * FROM changed
    UNION ALL
    SELECT l.* FROM documents_knowledge_file_leases l
    JOIN candidate c ON c.lease_id = l.lease_id
    WHERE l.lifecycle_state = 'STAGED' AND NOT EXISTS (SELECT 1 FROM changed)
)
SELECT lease_id, upload_id, actor_id, item_id, version_id,
       reference_id, source_id, expected_plaintext_bytes,
       expected_plaintext_sha256, filename, declared_mime,
       detected_mime, scan_signature_revision, scanned_at,
       lifecycle_state, expires_at, adopted_at, released_at
FROM result;
