SELECT lease_id, upload_id, actor_id, item_id, version_id,
       reference_id, source_id, expected_plaintext_bytes,
       expected_plaintext_sha256, filename, declared_mime,
       detected_mime, scan_signature_revision, scanned_at,
       lifecycle_state, expires_at, adopted_at, released_at
FROM documents_knowledge_file_leases
WHERE item_id = $1::uuid
  AND reference_id = $2::uuid AND source_id = $3::uuid
  AND lifecycle_state IN ('ADOPTED','RELEASED')
FOR UPDATE;
