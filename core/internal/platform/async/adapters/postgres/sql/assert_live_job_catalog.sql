SELECT NOT EXISTS (
    SELECT 1
    FROM platform.jobs AS job
    WHERE job.state IN ('queued', 'running')
      AND NOT EXISTS (
          SELECT 1
          FROM pg_catalog.generate_subscripts($1::text[], 1) AS policy(position)
          WHERE ($1::text[])[policy.position] = job.kind
            AND ($2::integer[])[policy.position] = job.payload_revision
            AND ($3::text[])[policy.position] = job.authority_domain
            AND ($4::text[])[policy.position] = job.effect_class
            AND ($5::integer[])[policy.position] = job.max_attempts
      )
);
