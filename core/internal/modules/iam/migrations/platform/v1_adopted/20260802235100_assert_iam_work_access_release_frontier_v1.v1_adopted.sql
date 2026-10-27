-- +goose Up
-- Work access authority owns tenant database generation 96.
-- Advance the platform frontier only after its sealed payload prerequisite exists.
SELECT 1 / CASE WHEN
    pg_catalog.current_setting('server_version_num')::integer
        BETWEEN 180000 AND 189999
    AND (
        SELECT NOT role_record.rolsuper
           AND NOT role_record.rolinherit
           AND NOT role_record.rolcreaterole
           AND NOT role_record.rolcreatedb
           AND NOT role_record.rolcanlogin
           AND NOT role_record.rolbypassrls
        FROM pg_catalog.pg_roles AS role_record
        WHERE role_record.rolname = 'harpy_runtime'
    )
    AND (
        SELECT NOT role_record.rolsuper
           AND NOT role_record.rolinherit
           AND NOT role_record.rolcreaterole
           AND NOT role_record.rolcreatedb
           AND role_record.rolcanlogin
           AND NOT role_record.rolbypassrls
        FROM pg_catalog.pg_roles AS role_record
        WHERE role_record.rolname = 'harpy_v2_runtime'
    )
THEN 1 ELSE 0 END
AS iam_work_access_release_frontier_assertion;
