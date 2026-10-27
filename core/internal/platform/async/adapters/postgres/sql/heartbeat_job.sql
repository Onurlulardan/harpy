WITH heartbeat_clock AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS observed_at
), heartbeat AS (
    UPDATE platform.jobs AS job
    SET lease_heartbeat_at = GREATEST(
            heartbeat_clock.observed_at,
            job.lease_heartbeat_at + interval '1 microsecond',
            job.updated_at + interval '1 microsecond'
        ),
        lease_until = GREATEST(
            heartbeat_clock.observed_at,
            job.lease_heartbeat_at + interval '1 microsecond',
            job.updated_at + interval '1 microsecond'
        ) + interval '60 seconds',
        updated_at = GREATEST(
            heartbeat_clock.observed_at,
            job.lease_heartbeat_at + interval '1 microsecond',
            job.updated_at + interval '1 microsecond'
        )
    FROM heartbeat_clock
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
      AND job.lease_until > heartbeat_clock.observed_at
    RETURNING
        job.lease_heartbeat_at,
        job.lease_until,
        job.cancel_requested_at IS NOT NULL AS cancel_requested
)
SELECT lease_heartbeat_at, lease_until, cancel_requested
FROM heartbeat;
