-- +goose Up
-- Tenant supersession chain hardening advances the release frontier to 82.
SELECT 1 / CASE WHEN
    pg_catalog.current_setting('server_version_num')::integer
        BETWEEN 180000 AND 189999
    AND (
        SELECT pg_catalog.count(*) = 6
        FROM platform.tenant_migration_supersessions AS supersession
        WHERE supersession.owner = 'system_resource_layout'
          AND supersession.scope = 'tenant'
          AND supersession.logical_version = 20260730120104
    )
    AND (
        SELECT pg_catalog.count(*) = 6
        FROM platform.tenant_migration_supersessions AS actual
        JOIN (VALUES
          ('fresh',
           'core/internal/modules/system_resource_layout/migrations/tenant/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql',
           '561757a212e2205e14091db2f9c05ab94f97d5a8cc67438d16c44a2ed4d0a8a2',
           'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql',
           '75ca56659160fe477fda1ddedec60389e378f410da3ef9e6b5dacfc885a98969',1),
          ('fresh',
           'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql',
           '75ca56659160fe477fda1ddedec60389e378f410da3ef9e6b5dacfc885a98969',
           'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v2/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql',
           '93110a1aba60b236a4c9fb96cc8df675ec0efac81766b6af391e69b217819b73',1),
          ('fresh',
           'core/internal/modules/system_resource_layout/migrations/tenant/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql',
           '561757a212e2205e14091db2f9c05ab94f97d5a8cc67438d16c44a2ed4d0a8a2',
           'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v2/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql',
           '93110a1aba60b236a4c9fb96cc8df675ec0efac81766b6af391e69b217819b73',2),
          ('v1_adopted',
           'core/internal/modules/system_resource_layout/migrations/tenant/v1_adopted/20260730120104_repair_installation_layout_publications_v1.v1_adopted.sql',
           '561757a212e2205e14091db2f9c05ab94f97d5a8cc67438d16c44a2ed4d0a8a2',
           'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v1_adopted/20260730120104_repair_installation_layout_publications_v1.v1_adopted.sql',
           '75ca56659160fe477fda1ddedec60389e378f410da3ef9e6b5dacfc885a98969',1),
          ('v1_adopted',
           'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v1_adopted/20260730120104_repair_installation_layout_publications_v1.v1_adopted.sql',
           '75ca56659160fe477fda1ddedec60389e378f410da3ef9e6b5dacfc885a98969',
           'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v2/v1_adopted/20260730120104_repair_installation_layout_publications_v1.v1_adopted.sql',
           '93110a1aba60b236a4c9fb96cc8df675ec0efac81766b6af391e69b217819b73',1),
          ('v1_adopted',
           'core/internal/modules/system_resource_layout/migrations/tenant/v1_adopted/20260730120104_repair_installation_layout_publications_v1.v1_adopted.sql',
           '561757a212e2205e14091db2f9c05ab94f97d5a8cc67438d16c44a2ed4d0a8a2',
           'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v2/v1_adopted/20260730120104_repair_installation_layout_publications_v1.v1_adopted.sql',
           '93110a1aba60b236a4c9fb96cc8df675ec0efac81766b6af391e69b217819b73',2)
        ) AS expected(
            origin,superseded_path,superseded_sha256,
            current_path,current_sha256,distance
        )
          ON actual.owner = 'system_resource_layout'
         AND actual.scope = 'tenant'
         AND actual.logical_version = 20260730120104
         AND actual.database_origin = expected.origin
         AND actual.superseded_path = expected.superseded_path
         AND actual.superseded_sha256 = expected.superseded_sha256
         AND actual.current_path = expected.current_path
         AND actual.current_sha256 = expected.current_sha256
         AND actual.supersession_distance = expected.distance
    )
    AND pg_catalog.to_regprocedure(
        'platform.platform_migrator_tenant_checkpoint_satisfies_v2(uuid,text,bigint,text,text,text)'
    ) IS NOT NULL
    AND pg_catalog.to_regprocedure(
        'platform.platform_migrator_verify_tenant_requirement_catalog_v4(uuid,uuid,uuid,text)'
    ) IS NOT NULL
    AND (
        SELECT pg_catalog.count(*) = 1
        FROM pg_catalog.pg_trigger AS trigger_record
        JOIN pg_catalog.pg_class AS relation
          ON relation.oid = trigger_record.tgrelid
        JOIN pg_catalog.pg_namespace AS namespace
          ON namespace.oid = relation.relnamespace
        JOIN pg_catalog.pg_proc AS routine
          ON routine.oid = trigger_record.tgfoid
        WHERE namespace.nspname = 'platform'
          AND relation.relname = 'tenant_migration_checkpoints'
          AND trigger_record.tgname =
              'tenant_migration_checkpoints_validate_transition_v4'
          AND trigger_record.tgenabled = 'A'
          AND NOT trigger_record.tgisinternal
          AND routine.proname =
              'platform_migrator_validate_checkpoint_transition_v4'
    )
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
AS tenant_supersession_chain_release_frontier_assertion;
