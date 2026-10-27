WITH selected AS MATERIALIZED (
    SELECT l.lease_id FROM documents_knowledge_file_leases l
    JOIN documents_items i ON i.id = l.item_id
    WHERE l.lifecycle_state IN ('PENDING','STAGED') AND l.expires_at <= $1
      AND i.lifecycle_state = 'ACTIVE'
    ORDER BY l.expires_at, l.lease_id LIMIT $2 FOR UPDATE OF l, i SKIP LOCKED
), changed AS (
    UPDATE documents_knowledge_file_leases l
    SET lifecycle_state = 'EXPIRED', expired_at = $1, updated_at = $1
    FROM selected s WHERE l.lease_id = s.lease_id
    RETURNING l.*
)
SELECT l.lease_id, l.upload_id, l.actor_id, l.item_id, l.version_id,
       l.reference_id, l.source_id, l.expected_plaintext_bytes,
       l.expected_plaintext_sha256, l.filename, l.declared_mime,
       l.detected_mime, l.scan_signature_revision, l.scanned_at,
       l.lifecycle_state, l.expires_at, l.adopted_at, l.released_at,
       i.row_version
FROM changed l JOIN documents_items i ON i.id = l.item_id
ORDER BY l.expires_at, l.lease_id;
