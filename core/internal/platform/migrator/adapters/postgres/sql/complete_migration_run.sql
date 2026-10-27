UPDATE platform.migration_runs
SET state = 'succeeded',
    finished_at = pg_catalog.statement_timestamp()
WHERE run_id = $1
  AND state = 'running';
