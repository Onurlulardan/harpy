UPDATE work_recurrence_rules
SET current_revision = $2, row_version = row_version + 1, updated_at = transaction_timestamp()
WHERE id = $1 AND current_revision = $3 AND lifecycle_state = 'ACTIVE'
