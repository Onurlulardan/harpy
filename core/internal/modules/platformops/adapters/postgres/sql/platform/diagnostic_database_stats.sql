SELECT pg_catalog.current_database(),
       pg_catalog.pg_database_size(pg_catalog.current_database())::text,
       pg_catalog.current_setting('server_version_num'),
       pg_catalog.current_setting('transaction_read_only')
FROM platform.platformops_recovery_signing_bundles AS anchor
RIGHT JOIN (VALUES (1)) AS singleton(value) ON false
LIMIT 1
