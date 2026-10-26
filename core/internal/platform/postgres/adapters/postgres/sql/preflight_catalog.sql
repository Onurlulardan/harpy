SELECT
    pg_catalog.current_setting('server_version_num')::integer,
    database.oid::bigint,
    database.datname,
    pg_catalog.pg_get_userbyid(database.datdba),
    pg_catalog.pg_encoding_to_char(database.encoding),
    CASE database.datlocprovider
        WHEN 'c' THEN 'libc'
        WHEN 'i' THEN 'icu'
        WHEN 'b' THEN 'builtin'
        ELSE database.datlocprovider::text
    END,
    database.datcollate,
    database.datctype,
    pg_catalog.current_setting('data_checksums'),
    pg_catalog.pg_is_in_recovery(),
    extension.extversion,
    namespace.nspname,
	pg_catalog.pg_get_userbyid(extension.extowner),
    control.system_identifier::text
FROM pg_catalog.pg_database AS database
CROSS JOIN pg_catalog.pg_control_system() AS control
JOIN pg_catalog.pg_extension AS extension ON extension.extname = 'vector'
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = extension.extnamespace
WHERE database.datname = pg_catalog.current_database();
