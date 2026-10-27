UPDATE work_operations
SET state = 'COMPLETED', aggregate_id = $2, aggregate_version = $3,
    response_status = 200, response_document = $4, completed_at = transaction_timestamp()
WHERE operation_id = $1 AND state = 'CLAIMED'
