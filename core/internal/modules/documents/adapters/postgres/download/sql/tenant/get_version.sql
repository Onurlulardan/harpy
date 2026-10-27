SELECT i.id, v.id, v.original_name, v.detected_mime,
       v.object_key, v.object_version, v.object_generation,
       v.plaintext_bytes, v.ciphertext_bytes, v.plaintext_sha256,
       v.ciphertext_sha256, v.key_id, v.wrapped_dek, v.created_at
FROM documents_items i
JOIN documents_versions v ON v.item_id = i.id
WHERE i.id = $1::uuid AND v.id = $2::uuid
  AND i.lifecycle_state IN ('ACTIVE','TRASHED','PURGE_COOLING')
  AND v.lifecycle_state = 'READY';
