UPDATE documents_forms_cleanup_targets
SET enqueued_at = $5::timestamptz
WHERE operation_id = $1::uuid AND item_id = $2::uuid
  AND purge_job_id = $3::uuid AND purge_run_at = $4::timestamptz
  AND enqueued_at IS NULL
RETURNING operation_id;
