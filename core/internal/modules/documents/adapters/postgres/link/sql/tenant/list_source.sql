SELECT id, item_id, source_owner, source_kind, source_id, source_version_id,
       usage, reference_epoch, lifecycle_state, created_at, released_at, false
FROM documents_links
WHERE source_owner = $1 AND source_kind = $2
  AND source_id = $3::uuid AND source_version_id = $4::uuid
  AND usage = 'LIFECYCLE' AND lifecycle_state = 'ACTIVE'
ORDER BY item_id, id
LIMIT $5
FOR UPDATE;
