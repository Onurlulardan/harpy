SELECT
    (u.cancel_requested_at IS NULL AND
     ((u.expected_current_version_id IS NULL AND i.id IS NULL) OR
      i.current_version_id = u.expected_current_version_id)) AS ready,
    (u.cancel_requested_at IS NULL AND NOT
     ((u.expected_current_version_id IS NULL AND i.id IS NULL) OR
      i.current_version_id = u.expected_current_version_id)) AS conflict,
    (u.cancel_requested_at IS NOT NULL) AS cancelled,
    i.current_version_id
FROM documents_uploads u
JOIN documents_upload_attempts a
  ON a.upload_id = u.id AND a.attempt_generation = u.attempt_generation
LEFT JOIN LATERAL (
    SELECT item.id, item.current_version_id
    FROM documents_items item WHERE item.id = u.document_id
    FOR UPDATE
) i ON true
WHERE u.id = $1::uuid AND u.actor_id = $2::uuid
  AND u.attempt_generation = $3 AND u.lifecycle_state = 'SCANNING'
  AND a.lifecycle_state = 'EFFECT_RECORDED' AND a.object_version = $4
FOR UPDATE OF u;
