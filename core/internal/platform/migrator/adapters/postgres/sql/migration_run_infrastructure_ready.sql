SELECT pg_catalog.to_regclass('platform.migration_runs') IS NOT NULL
   AND pg_catalog.to_regprocedure(
       'platform.platform_migrator_validate_run_transition()'
   ) IS NOT NULL
   AND EXISTS (
       SELECT 1
       FROM pg_catalog.pg_trigger AS trigger_row
       JOIN pg_catalog.pg_class AS relation_row
         ON relation_row.oid = trigger_row.tgrelid
       JOIN pg_catalog.pg_namespace AS namespace_row
         ON namespace_row.oid = relation_row.relnamespace
       WHERE namespace_row.nspname = 'platform'
         AND relation_row.relname = 'migration_runs'
         AND trigger_row.tgname = 'migration_runs_validate_transition'
         AND NOT trigger_row.tgisinternal
   );
