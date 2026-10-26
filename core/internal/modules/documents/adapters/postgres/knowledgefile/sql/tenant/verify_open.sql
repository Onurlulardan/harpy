SELECT l.lease_id, l.upload_id, l.actor_id, l.item_id, l.version_id,
       l.reference_id, l.source_id, l.expected_plaintext_bytes,
       l.expected_plaintext_sha256, l.filename, l.declared_mime,
       l.detected_mime, l.scan_signature_revision, l.scanned_at,
       l.lifecycle_state, l.expires_at, l.adopted_at, l.released_at
FROM documents_knowledge_file_leases l
JOIN documents_items i ON i.id = l.item_id
JOIN documents_versions v ON v.id = l.version_id AND v.item_id = i.id
JOIN documents_links r ON r.id = l.reference_id AND r.item_id = i.id
WHERE l.actor_id = $1::uuid AND l.item_id = $2::uuid
  AND l.version_id = $3::uuid AND l.lifecycle_state = 'ADOPTED'
  AND i.lifecycle_state = 'ACTIVE' AND i.current_version_id = v.id
  AND v.lifecycle_state = 'READY' AND v.plaintext_bytes = l.expected_plaintext_bytes
  AND v.plaintext_sha256 = l.expected_plaintext_sha256
  AND r.lifecycle_state = 'ACTIVE' AND r.source_owner = 'AI'
  AND r.source_kind = 'knowledge_document' AND r.source_id = l.source_id
  AND r.source_version_id = l.source_id AND r.usage = 'LIFECYCLE';
