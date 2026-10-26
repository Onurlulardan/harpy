WITH failed AS (
    UPDATE platform.jobs AS job
    SET state = $10,
    run_at = CASE
            WHEN $10::text = 'queued' THEN
                $15::timestamptz + ($14::bigint * interval '1 microsecond')
            ELSE job.run_at
        END,
        lease_owner_id = NULL,
        lease_heartbeat_at = NULL,
        lease_until = NULL,
        last_error_class = $11,
        last_error_code = $12,
        last_error_sha256 = $13,
        retain_until = CASE
        WHEN $10::text = 'dead' THEN $15::timestamptz + interval '365 days'
        WHEN $10::text = 'cancelled' THEN $15::timestamptz + interval '90 days'
        ELSE NULL
    END,
        updated_at = $15::timestamptz,
        completed_at = CASE
            WHEN $10::text = 'queued' THEN NULL
            ELSE $15::timestamptz
        END
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
      AND job.effect_started_at IS NOT DISTINCT FROM $16::timestamptz
      AND (
          ($10::text = 'queued' AND $11::text IN ('transient', 'conflict') AND
              job.attempts < job.max_attempts AND $14::bigint >= 0 AND
              ($14::bigint * interval '1 microsecond') <= LEAST(
                  interval '5 seconds' * pg_catalog.power(
                      2::double precision,
                      (job.attempts - 1)::double precision
                  ),
                  interval '15 minutes'
              )) OR
          ($10::text = 'dead' AND (
              $11::text = 'permanent' OR
              ($11::text IN ('transient', 'conflict') AND
                  job.attempts = job.max_attempts)
          ) AND $14::bigint = 0) OR
          ($10::text = 'cancelled' AND $11::text = 'cancelled' AND
              $14::bigint = 0)
      )
    RETURNING
        job.updated_at,
        CASE WHEN job.state = 'queued' THEN job.run_at END AS run_at
)
SELECT updated_at, run_at
FROM failed;
