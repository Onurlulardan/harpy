SELECT pg_catalog.convert_to(pg_catalog.jsonb_build_object(
    'id', i.id::text, 'folder_id', i.folder_id::text,
    'owner_id', i.owner_id::text, 'current_version_id', i.current_version_id::text,
    'state', i.lifecycle_state, 'row_version', i.row_version,
    'description', i.description, 'tags', i.tags,
    'trashed_at', i.trashed_at, 'retention_ends_at', i.retention_ends_at,
    'purge_requested_at', i.purge_requested_at, 'purge_starts_at', i.purge_starts_at,
    'created_at', i.created_at, 'updated_at', i.updated_at,
    'current', pg_catalog.jsonb_build_object(
        'id', v.id::text, 'number', v.version_number,
        'original_name', v.original_name, 'mime', v.detected_mime,
        'file_kind', v.file_kind, 'plaintext_bytes', v.plaintext_bytes,
        'plaintext_sha256', pg_catalog.encode(v.plaintext_sha256, 'hex'),
        'created_at', v.created_at
    )
)::text, 'UTF8')
FROM documents_items i
JOIN documents_versions v ON v.id = i.current_version_id
WHERE i.folder_id = $1::uuid AND i.lifecycle_state <> 'PURGED'
  AND v.lifecycle_state = 'READY'
  AND ($2::timestamptz IS NULL OR (i.updated_at, i.id) < ($2, $3::uuid))
ORDER BY i.updated_at DESC, i.id DESC
LIMIT $4;
