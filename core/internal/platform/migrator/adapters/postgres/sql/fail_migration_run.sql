UPDATE platform.migration_runs
SET state = 'failed',
    finished_at = pg_catalog.statement_timestamp(),
    failure_code = $2
WHERE run_id = $1
  AND state = 'running';
