WITH anchor AS (
    SELECT display_name, id FROM documents_folders WHERE id = $2::uuid
)
SELECT pg_catalog.jsonb_build_object(
    'id', folder.id::text, 'parent_id', folder.parent_id::text,
    'owner_id', folder.owner_id::text, 'display_name', folder.display_name,
    'state', folder.lifecycle_state, 'row_version', folder.row_version,
    'created_at', folder.created_at, 'updated_at', folder.updated_at,
    'trashed_at', folder.trashed_at
)
FROM documents_folders folder
WHERE folder.parent_id IS NOT DISTINCT FROM $1::uuid
  AND folder.lifecycle_state = 'ACTIVE'
  AND ($2::uuid IS NULL OR (folder.display_name COLLATE "C", folder.id) > (
      SELECT display_name COLLATE "C", id FROM anchor
  ))
ORDER BY folder.display_name COLLATE "C", folder.id
LIMIT $3;
