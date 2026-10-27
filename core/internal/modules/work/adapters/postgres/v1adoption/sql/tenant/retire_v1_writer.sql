UPDATE work_v1_writer_fence
SET state = 'RETIRED', reverse_reconciliation_id = NULL,
    updated_by = $2, updated_at = transaction_timestamp()
WHERE singleton AND state = 'V2_FENCED' AND run_id = $1
