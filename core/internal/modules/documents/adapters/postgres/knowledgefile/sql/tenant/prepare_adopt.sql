SELECT lease_id, upload_id, actor_id, item_id, version_id,
       reference_id, source_id, expected_plaintext_bytes,
       expected_plaintext_sha256, filename, declared_mime,
       detected_mime, scan_signature_revision, scanned_at,
       lifecycle_state, expires_at, adopted_at, released_at
FROM documents_knowledge_file_leases
WHERE lease_id = $1::uuid AND actor_id = $2::uuid
  AND item_id = $3::uuid AND version_id = $4::uuid
  AND (lifecycle_state = 'STAGED' AND expires_at > $7
       OR lifecycle_state = 'ADOPTED'
          AND reference_id = $5::uuid AND source_id = $6::uuid)
FOR UPDATE;
