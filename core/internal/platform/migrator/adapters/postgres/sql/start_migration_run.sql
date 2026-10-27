UPDATE platform.migration_runs
SET state = 'running',
    started_at = pg_catalog.statement_timestamp()
WHERE run_id = $1
  AND state = 'planned';
