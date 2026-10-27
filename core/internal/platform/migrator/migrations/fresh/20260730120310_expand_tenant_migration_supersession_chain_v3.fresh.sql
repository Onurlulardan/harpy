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
    '596819b7a320d41d283ea4d88846566a16bcbe6e6739701b64af77bca93d524d',
    'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v3/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql',
    'cef35a91f7e9cffebc4a48a3c26d69b5edb90cbd27afae1e67dc606c58309972',
    1
),
(
    'system_resource_layout','tenant',20260730120104,'fresh',
    'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql',
    '9ecca5ceb2b0453956a81f46e46d2bf633890ee5e8c212002b33b35d73faa575',
    'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v3/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql',
    'cef35a91f7e9cffebc4a48a3c26d69b5edb90cbd27afae1e67dc606c58309972',
    2
),
(
    'system_resource_layout','tenant',20260730120104,'fresh',
    'core/internal/modules/system_resource_layout/migrations/tenant/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql',
    '561757a212e2205e14091db2f9c05ab94f97d5a8cc67438d16c44a2ed4d0a8a2',
    'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v3/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql',
    'cef35a91f7e9cffebc4a48a3c26d69b5edb90cbd27afae1e67dc606c58309972',
    3
),
(
    'system_resource_layout','tenant',20260730120104,'v1_adopted',
    'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v2/v1_adopted/20260730120104_repair_installation_layout_publications_v1.v1_adopted.sql',
    '596819b7a320d41d283ea4d88846566a16bcbe6e6739701b64af77bca93d524d',
    'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v3/v1_adopted/20260730120104_repair_installation_layout_publications_v1.v1_adopted.sql',
    'cef35a91f7e9cffebc4a48a3c26d69b5edb90cbd27afae1e67dc606c58309972',
    1
),
(
    'system_resource_layout','tenant',20260730120104,'v1_adopted',
    'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v1_adopted/20260730120104_repair_installation_layout_publications_v1.v1_adopted.sql',
    '9ecca5ceb2b0453956a81f46e46d2bf633890ee5e8c212002b33b35d73faa575',
    'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v3/v1_adopted/20260730120104_repair_installation_layout_publications_v1.v1_adopted.sql',
    'cef35a91f7e9cffebc4a48a3c26d69b5edb90cbd27afae1e67dc606c58309972',
    2
),
(
    'system_resource_layout','tenant',20260730120104,'v1_adopted',
    'core/internal/modules/system_resource_layout/migrations/tenant/v1_adopted/20260730120104_repair_installation_layout_publications_v1.v1_adopted.sql',
    '561757a212e2205e14091db2f9c05ab94f97d5a8cc67438d16c44a2ed4d0a8a2',
    'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v3/v1_adopted/20260730120104_repair_installation_layout_publications_v1.v1_adopted.sql',
    'cef35a91f7e9cffebc4a48a3c26d69b5edb90cbd27afae1e67dc606c58309972',
    3
);
