-- +goose Up
ALTER TABLE platform.tenant_migration_supersessions
    DROP CONSTRAINT tenant_migration_supersessions_distance_check,
    ADD CONSTRAINT tenant_migration_supersessions_distance_check CHECK (
        supersession_distance IN (1,2,3)
        AND position('/'||database_origin||'/' IN superseded_path) > 0
        AND position('/'||database_origin||'/' IN current_path) > 0
        AND (
            supersession_distance = 1 AND (
                position('/supersessions/' IN superseded_path) = 0
                AND position('/supersessions/'||database_origin||'/' IN current_path) > 0
                AND position('/supersessions/v2/' IN current_path) = 0
                AND position('/supersessions/v3/' IN current_path) = 0
                OR position('/supersessions/'||database_origin||'/' IN superseded_path) > 0
                AND position('/supersessions/v2/' IN superseded_path) = 0
                AND position('/supersessions/v3/' IN superseded_path) = 0
                AND position('/supersessions/v2/' IN current_path) > 0
                OR position('/supersessions/v2/' IN superseded_path) > 0
                AND position('/supersessions/v3/' IN current_path) > 0
            )
            OR supersession_distance = 2 AND (
                position('/supersessions/' IN superseded_path) = 0
                AND position('/supersessions/v2/' IN current_path) > 0
                OR position('/supersessions/'||database_origin||'/' IN superseded_path) > 0
                AND position('/supersessions/v2/' IN superseded_path) = 0
                AND position('/supersessions/v3/' IN current_path) > 0
            )
            OR supersession_distance = 3
            AND position('/supersessions/' IN superseded_path) = 0
            AND position('/supersessions/v3/' IN current_path) > 0
        )
    );

INSERT INTO platform.tenant_migration_supersessions (
    owner,scope,logical_version,database_origin,
    superseded_path,superseded_sha256,current_path,current_sha256,
    supersession_distance
) VALUES
(
    'system_resource_layout','tenant',20260730120104,'fresh',
    'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v2/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql',
    '93110a1aba60b236a4c9fb96cc8df675ec0efac81766b6af391e69b217819b73',
    'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v3/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql',
    'a6523dc1689b2bfa94ffed63e432a6d738cf0a86b3e4fb09e0c160adf1730724',
    1
),
(
    'system_resource_layout','tenant',20260730120104,'fresh',
    'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql',
    '75ca56659160fe477fda1ddedec60389e378f410da3ef9e6b5dacfc885a98969',
    'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v3/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql',
    'a6523dc1689b2bfa94ffed63e432a6d738cf0a86b3e4fb09e0c160adf1730724',
    2
),
(
    'system_resource_layout','tenant',20260730120104,'fresh',
    'core/internal/modules/system_resource_layout/migrations/tenant/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql',
    '561757a212e2205e14091db2f9c05ab94f97d5a8cc67438d16c44a2ed4d0a8a2',
    'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v3/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql',
    'a6523dc1689b2bfa94ffed63e432a6d738cf0a86b3e4fb09e0c160adf1730724',
    3
),
(
    'system_resource_layout','tenant',20260730120104,'v1_adopted',
    'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v2/v1_adopted/20260730120104_repair_installation_layout_publications_v1.v1_adopted.sql',
    '93110a1aba60b236a4c9fb96cc8df675ec0efac81766b6af391e69b217819b73',
    'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v3/v1_adopted/20260730120104_repair_installation_layout_publications_v1.v1_adopted.sql',
    'a6523dc1689b2bfa94ffed63e432a6d738cf0a86b3e4fb09e0c160adf1730724',
    1
),
(
    'system_resource_layout','tenant',20260730120104,'v1_adopted',
    'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v1_adopted/20260730120104_repair_installation_layout_publications_v1.v1_adopted.sql',
    '75ca56659160fe477fda1ddedec60389e378f410da3ef9e6b5dacfc885a98969',
    'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v3/v1_adopted/20260730120104_repair_installation_layout_publications_v1.v1_adopted.sql',
    'a6523dc1689b2bfa94ffed63e432a6d738cf0a86b3e4fb09e0c160adf1730724',
    2
),
(
    'system_resource_layout','tenant',20260730120104,'v1_adopted',
    'core/internal/modules/system_resource_layout/migrations/tenant/v1_adopted/20260730120104_repair_installation_layout_publications_v1.v1_adopted.sql',
    '561757a212e2205e14091db2f9c05ab94f97d5a8cc67438d16c44a2ed4d0a8a2',
    'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v3/v1_adopted/20260730120104_repair_installation_layout_publications_v1.v1_adopted.sql',
    'a6523dc1689b2bfa94ffed63e432a6d738cf0a86b3e4fb09e0c160adf1730724',
    3
);
