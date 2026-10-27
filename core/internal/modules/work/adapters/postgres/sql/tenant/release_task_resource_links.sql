UPDATE work_resource_links
SET state = 'RELEASED', released_at = transaction_timestamp(),
    row_version = row_version + 1
WHERE work_resource_kind = 'TASK'
  AND work_resource_id = $1
  AND state = 'ACTIVE'
