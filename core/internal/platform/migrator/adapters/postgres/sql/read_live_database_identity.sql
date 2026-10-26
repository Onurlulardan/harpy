SELECT pg_catalog.current_database(),
       owner_role.rolname,
       pg_catalog.current_setting('server_encoding'),
       CASE database_row.datlocprovider
           WHEN 'c' THEN 'libc'
           WHEN 'i' THEN 'icu'
           WHEN 'b' THEN 'builtin'
           ELSE database_row.datlocprovider::text
       END,
       database_row.datcollate,
       database_row.datctype,
       control.system_identifier::text,
       database_row.oid,
       pg_catalog.current_setting('data_checksums') = 'on',
       pg_catalog.current_setting('server_version_num')::integer,
       session_user,
       current_user,
       pg_catalog.current_setting('search_path'),
       pg_catalog.current_setting('transaction_read_only') = 'on',
       pg_catalog.pg_is_in_recovery()
FROM pg_catalog.pg_database AS database_row
JOIN pg_catalog.pg_roles AS owner_role ON owner_role.oid = database_row.datdba
CROSS JOIN pg_catalog.pg_control_system() AS control
WHERE database_row.datname = pg_catalog.current_database();
