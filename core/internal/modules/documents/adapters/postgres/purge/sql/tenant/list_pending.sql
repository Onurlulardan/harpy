SELECT d.id, d.target_id, d.object_key, d.object_version,
       d.object_generation, d.plaintext_bytes, d.ciphertext_bytes
FROM documents_blob_deletions d
JOIN documents_versions v ON v.id = d.target_id
WHERE d.target_kind = 'VERSION' AND v.item_id = $1::uuid
  AND d.lifecycle_state <> 'CONFIRMED'
  AND ($2::uuid IS NULL OR d.id > $2::uuid)
ORDER BY d.id
LIMIT $3;
