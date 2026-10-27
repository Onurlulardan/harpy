-- +goose Up
-- Designer menu JSON storage owns database generation 97.
-- Advance the platform frontier before the tenant catalog_menu_configs table.
SELECT 1 / CASE WHEN
    pg_catalog.current_setting('server_version_num')::integer
        BETWEEN 180000 AND 189999
    AND pg_catalog.to_regclass(
        'platform.catalog_package_owner_payloads'
    ) IS NOT NULL
    AND (
        SELECT NOT role_record.rolsuper
           AND NOT role_record.rolinherit
           AND NOT role_record.rolcreaterole
           AND NOT role_record.rolcreatedb
           AND NOT role_record.rolcanlogin
           AND NOT role_record.rolbypassrls
        FROM pg_catalog.pg_roles AS role_record
        WHERE role_record.rolname = 'forgecrud_runtime'
    )
    AND (
        SELECT NOT role_record.rolsuper
           AND NOT role_record.rolinherit
           AND NOT role_record.rolcreaterole
           AND NOT role_record.rolcreatedb
           AND role_record.rolcanlogin
           AND NOT role_record.rolbypassrls
        FROM pg_catalog.pg_roles AS role_record
        WHERE role_record.rolname = 'forgecrud_v2_runtime'
    )
THEN 1 ELSE 0 END
AS catalog_menu_config_release_frontier_assertion;
