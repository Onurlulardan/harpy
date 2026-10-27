SELECT (pg_catalog.to_regclass('platform.platformops_recovery_sets') IS NOT NULL)::text,
       COALESCE((
           SELECT pg_catalog.max(generation)::text
           FROM platform.platformops_recovery_signing_bundles
       ), '0'),
       COALESCE((
           SELECT pg_catalog.count(*)::text
           FROM platform.platformops_recovery_sets WHERE state = 'protected'
       ), '0'),
       pg_catalog.current_setting('transaction_read_only')
