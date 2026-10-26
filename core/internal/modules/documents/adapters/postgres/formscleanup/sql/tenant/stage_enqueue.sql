WITH changed AS (
    UPDATE documents_forms_cleanup_targets
    SET purge_job_id = $3::uuid, purge_run_at = $4::timestamptz
    WHERE operation_id = $1::uuid AND item_id = $2::uuid
      AND purge_job_id IS NULL AND enqueued_at IS NULL
    RETURNING item_id, purge_job_id, purge_run_at
)
SELECT item_id, purge_job_id, purge_run_at FROM changed
UNION ALL
SELECT item_id, purge_job_id, purge_run_at
FROM documents_forms_cleanup_targets
WHERE operation_id = $1::uuid AND item_id = $2::uuid
  AND purge_job_id IS NOT NULL AND enqueued_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM changed);
