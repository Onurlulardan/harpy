WITH changed AS (
    UPDATE documents_knowledge_file_leases
    SET lifecycle_state = 'RELEASED', released_at = $4, updated_at = $4
    WHERE item_id = $1::uuid
      AND reference_id = $2::uuid AND source_id = $3::uuid
      AND lifecycle_state = 'ADOPTED'
    RETURNING *
), result AS (
    SELECT * FROM changed
    UNION ALL
    SELECT l.* FROM documents_knowledge_file_leases l
    WHERE l.item_id = $1::uuid
      AND l.reference_id = $2::uuid AND l.source_id = $3::uuid
      AND l.lifecycle_state = 'RELEASED' AND NOT EXISTS (SELECT 1 FROM changed)
)
SELECT lease_id, upload_id, actor_id, item_id, version_id,
       reference_id, source_id, expected_plaintext_bytes,
       expected_plaintext_sha256, filename, declared_mime,
       detected_mime, scan_signature_revision, scanned_at,
       lifecycle_state, expires_at, adopted_at, released_at
FROM result;
