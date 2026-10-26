UPDATE work_calendar_items
SET state = 'REVOKED', updated_at = transaction_timestamp()
WHERE source_owner = $1 AND source_kind = $2 AND source_id = $3
