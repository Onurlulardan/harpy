WITH lock_key AS (
    SELECT pg_catalog.hashtextextended(
        'harpy:v2:migrate:tenant:' || $4, 0
    ) AS value
)
SELECT pg_catalog.pg_backend_pid() = $1,
       pg_catalog.current_setting('search_path') = $2,
       pg_catalog.current_schema() = $3,
       EXISTS (
           SELECT 1
           FROM pg_catalog.pg_locks AS held, lock_key
           WHERE held.pid = pg_catalog.pg_backend_pid()
             AND held.locktype = 'advisory'
             AND held.granted
             AND held.classid = ((lock_key.value >> 32) & 4294967295)::oid
             AND held.objid = (lock_key.value & 4294967295)::oid
             AND held.objsubid = 1
       );
