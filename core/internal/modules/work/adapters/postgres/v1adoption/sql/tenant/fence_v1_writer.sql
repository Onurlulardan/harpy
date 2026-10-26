UPDATE work_v1_writer_fence
SET state = 'V2_FENCED', run_id = $1, generation = $2,
    reverse_reconciliation_id = NULL, updated_by = $3, updated_at = transaction_timestamp()
WHERE singleton AND (
    state = 'V1_ACTIVE' OR
    (state = 'V2_FENCED' AND run_id = $1 AND generation = $2)
)
