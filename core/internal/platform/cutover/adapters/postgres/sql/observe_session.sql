SELECT
    pg_catalog.pg_backend_pid(),
    activity.backend_start,
    SESSION_USER,
    pg_catalog.current_setting('application_name'),
    control.system_identifier::text,
    database_row.oid::bigint,
    pg_catalog.current_database()
FROM pg_catalog.pg_stat_activity AS activity
CROSS JOIN pg_catalog.pg_control_system() AS control
JOIN pg_catalog.pg_database AS database_row
  ON database_row.datname = pg_catalog.current_database()
WHERE activity.pid = pg_catalog.pg_backend_pid()
