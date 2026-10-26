SELECT pg_catalog.jsonb_build_object(
    'id', id::text, 'parent_id', parent_id::text, 'owner_id', owner_id::text,
    'display_name', display_name, 'state', lifecycle_state,
    'row_version', row_version, 'created_at', created_at,
    'updated_at', updated_at, 'trashed_at', trashed_at
)
FROM documents_folders
WHERE id = $1::uuid;
