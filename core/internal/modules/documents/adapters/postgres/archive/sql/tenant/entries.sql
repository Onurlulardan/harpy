SELECT e.document_id, e.version_id, e.entry_order, e.plaintext_bytes,
       e.plaintext_sha256, v.original_name, v.object_key, v.object_version,
       v.object_generation, v.ciphertext_bytes, v.ciphertext_sha256,
       v.key_id, v.wrapped_dek
FROM documents_archive_entries e
JOIN documents_versions v ON v.id = e.version_id
WHERE e.archive_id = $1::uuid
ORDER BY e.entry_order;
