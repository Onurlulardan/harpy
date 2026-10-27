UPDATE work_v1_writer_fence
SET state = 'V1_REOPENED', reverse_reconciliation_id = $2,
    updated_by = $3, updated_at = transaction_timestamp()
WHERE singleton AND state = 'V2_FENCED' AND run_id = $1
