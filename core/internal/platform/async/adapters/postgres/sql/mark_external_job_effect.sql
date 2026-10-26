WITH effect_clock AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS observed_at
), marked AS (
    UPDATE platform.jobs AS job
    SET effect_started_at = GREATEST(
            effect_clock.observed_at,
            job.updated_at + interval '1 microsecond'
        ),
        updated_at = GREATEST(
            effect_clock.observed_at,
            job.updated_at + interval '1 microsecond'
        )
    FROM effect_clock
    WHERE job.scope_kind = $1
      AND job.scope_id = $2
      AND job.authority_group_id = $3
      AND job.authority_group_generation = $4
      AND job.authority_domain = $5
      AND job.cutover_generation = $6
      AND job.id = $7
      AND job.state = 'running'
      AND job.effect_class = 'external'
      AND job.effect_started_at IS NULL
      AND job.cancel_requested_at IS NULL
      AND job.lease_owner_id = $8
      AND job.lease_generation = $9
      AND job.lease_until > effect_clock.observed_at
    RETURNING job.effect_started_at
)
SELECT effect_started_at
FROM marked;
