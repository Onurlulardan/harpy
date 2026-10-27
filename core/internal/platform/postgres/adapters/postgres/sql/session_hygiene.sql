SELECT
    session_user,
    current_user,
    pg_catalog.current_setting('application_name'),
    pg_catalog.current_setting('TimeZone'),
    pg_catalog.current_setting('search_path'),
    pg_catalog.current_setting('client_encoding'),
    pg_catalog.current_setting('session_replication_role'),
    pg_catalog.current_setting('row_security'),
    pg_catalog.pg_my_temp_schema()::bigint,
    NOT pg_catalog.pg_is_in_recovery(),
    NOT EXISTS (
        SELECT 1 FROM pg_catalog.pg_locks
        WHERE pid = pg_catalog.pg_backend_pid() AND locktype = 'advisory' AND granted
    ),
    NOT EXISTS (SELECT 1 FROM pg_catalog.pg_listening_channels());
