UPDATE work_inbox_items
SET state = 'REVOKED', permission_revision = $2, updated_at = transaction_timestamp()
WHERE id = $1
