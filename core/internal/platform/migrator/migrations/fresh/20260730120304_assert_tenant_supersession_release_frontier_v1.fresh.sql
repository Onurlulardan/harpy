-- +goose Up
-- Tenant checkpoint supersession hardening advances the release frontier to 81.
SELECT 1 / CASE WHEN
    pg_catalog.current_setting('server_version_num')::integer
        BETWEEN 180000 AND 189999
    AND (
        SELECT pg_catalog.count(*) = 2
        FROM platform.tenant_migration_supersessions AS supersession
        WHERE supersession.owner = 'system_resource_layout'
          AND supersession.scope = 'tenant'
          AND supersession.logical_version = 20260730120104
          AND supersession.superseded_sha256 =
              '561757a212e2205e14091db2f9c05ab94f97d5a8cc67438d16c44a2ed4d0a8a2'
          AND supersession.current_sha256 =
              '9ecca5ceb2b0453956a81f46e46d2bf633890ee5e8c212002b33b35d73faa575'
    )
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
AS tenant_supersession_release_frontier_assertion;
