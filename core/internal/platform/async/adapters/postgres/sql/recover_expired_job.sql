UPDATE platform.jobs AS job
SET state = $10::text,
    run_at = CASE WHEN $10::text = 'queued' THEN $14::timestamptz ELSE job.run_at END,
    lease_owner_id = NULL,
    lease_heartbeat_at = NULL,
    lease_until = NULL,
    last_error_class = $11::text,
    last_error_code = $12::text,
    last_error_sha256 = $13::text,
    retain_until = CASE
        WHEN $10::text = 'dead' THEN $14::timestamptz + interval '365 days'
        WHEN $10::text IN ('succeeded', 'cancelled') THEN $14::timestamptz + interval '90 days'
        ELSE NULL
    END,
    updated_at = $14::timestamptz,
    completed_at = CASE WHEN $10::text = 'queued' THEN NULL ELSE $14::timestamptz END
WHERE job.scope_kind = $1
  AND job.scope_id = $2
  AND job.authority_group_id = $3
  AND job.authority_group_generation = $4
  AND job.authority_domain = $5
  AND job.cutover_generation = $6
  AND job.id = $7
  AND job.state = 'running'
  AND job.lease_owner_id = $8
  AND job.lease_generation = $9
  AND job.lease_until <= $14::timestamptz
  AND (
      ($10::text = 'queued' AND $11::text IN ('transient', 'conflict') AND
          $12::text IS NOT NULL AND $13::text IS NOT NULL AND
          job.attempts < job.max_attempts) OR
      ($10::text = 'succeeded' AND $11::text IS NULL AND
          $12::text IS NULL AND $13::text IS NULL) OR
      ($10::text = 'dead' AND (
          ($11::text IN ('transient', 'permanent', 'conflict') AND
              $12::text IS NOT NULL AND $13::text IS NOT NULL) OR
          ($11::text IS NULL AND $12::text = 'JOB_OUTCOME_UNKNOWN' AND
              $13::text IS NULL)
      )) OR
      ($10::text = 'cancelled' AND $11::text = 'cancelled' AND
          $12::text IS NOT NULL AND $13::text IS NOT NULL)
  );
