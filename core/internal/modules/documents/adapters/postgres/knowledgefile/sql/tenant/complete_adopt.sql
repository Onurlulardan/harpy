WITH changed AS (
    UPDATE documents_knowledge_file_leases
    SET reference_id = $5::uuid, source_id = $6::uuid,
        lifecycle_state = 'ADOPTED', adopted_at = $7, updated_at = $7
    WHERE lease_id = $1::uuid AND actor_id = $2::uuid
      AND item_id = $3::uuid AND version_id = $4::uuid
      AND lifecycle_state = 'STAGED' AND expires_at > $7
    RETURNING *
), result AS (
    SELECT * FROM changed
    UNION ALL
    SELECT l.* FROM documents_knowledge_file_leases l
    WHERE l.lease_id = $1::uuid AND l.actor_id = $2::uuid
      AND l.item_id = $3::uuid AND l.version_id = $4::uuid
      AND l.reference_id = $5::uuid AND l.source_id = $6::uuid
      AND l.lifecycle_state = 'ADOPTED' AND NOT EXISTS (SELECT 1 FROM changed)
)
SELECT lease_id, upload_id, actor_id, item_id, version_id,
       reference_id, source_id, expected_plaintext_bytes,
       expected_plaintext_sha256, filename, declared_mime,
       detected_mime, scan_signature_revision, scanned_at,
       lifecycle_state, expires_at, adopted_at, released_at
FROM result;
