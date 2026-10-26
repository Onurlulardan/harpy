SELECT id, item_id, source_owner, source_kind, source_id, source_version_id,
       usage, reference_epoch, lifecycle_state, created_at, released_at, false
FROM documents_links
WHERE item_id = $1::uuid AND lifecycle_state = 'ACTIVE' AND usage = 'LIFECYCLE'
  AND ($2::uuid IS NULL OR id > $2::uuid)
ORDER BY id
LIMIT $3;
