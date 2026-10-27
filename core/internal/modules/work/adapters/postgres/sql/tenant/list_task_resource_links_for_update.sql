SELECT id, work_resource_id, target_owner, target_kind, target_id,
       target_version_id, target_reference_epoch, permission_revision, row_version
FROM work_resource_links
WHERE work_resource_kind = 'TASK'
  AND work_resource_id = $1
  AND state = 'ACTIVE'
ORDER BY CASE target_kind WHEN 'FORM' THEN 0 WHEN 'FORM_RECORD' THEN 1 ELSE 2 END,
         target_kind, id
LIMIT 3
FOR UPDATE
