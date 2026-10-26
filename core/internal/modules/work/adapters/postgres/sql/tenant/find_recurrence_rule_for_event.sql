SELECT id, current_revision
FROM work_recurrence_rules
WHERE owner_event_id = $1 AND lifecycle_state = 'ACTIVE'
FOR UPDATE
