SELECT v.id, v.version_number, v.original_name, v.detected_mime,
       v.file_kind, v.plaintext_bytes, v.plaintext_sha256, v.created_at
FROM documents_versions v
JOIN documents_items i ON i.id = v.item_id
WHERE v.item_id = $1::uuid AND v.lifecycle_state = 'READY'
  AND i.lifecycle_state <> 'PURGED' AND v.version_number > $2
ORDER BY v.version_number
LIMIT $3;
