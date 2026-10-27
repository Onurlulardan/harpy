-- +goose Up
CREATE TABLE platform.tenant_migration_supersessions (
    owner text NOT NULL CHECK (owner ~ '^[a-z][a-z0-9_-]{0,63}$'),
    scope text NOT NULL CHECK (scope = 'tenant'),
    logical_version bigint NOT NULL
        CHECK (logical_version BETWEEN 20000101000000 AND 99991231235959),
    database_origin text NOT NULL
        CHECK (database_origin IN ('fresh','v1_adopted')),
    superseded_path text NOT NULL,
    superseded_sha256 text NOT NULL
        CHECK (superseded_sha256 ~ '^[0-9a-f]{64}$'),
    current_path text NOT NULL,
    current_sha256 text NOT NULL
        CHECK (current_sha256 ~ '^[0-9a-f]{64}$'),
    registered_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    CONSTRAINT tenant_migration_supersessions_pkey
        PRIMARY KEY (owner,scope,logical_version,database_origin),
    CONSTRAINT tenant_migration_supersessions_superseded_unique
        UNIQUE (database_origin,superseded_path,superseded_sha256),
    CONSTRAINT tenant_migration_supersessions_current_unique
        UNIQUE (database_origin,current_path,current_sha256),
    CONSTRAINT tenant_migration_supersessions_paths_check CHECK (
        superseded_path ~ '^[a-z0-9][a-z0-9_./-]*\.sql$'
        AND current_path ~ '^[a-z0-9][a-z0-9_./-]*\.sql$'
        AND position('..' IN superseded_path) = 0
        AND position('//' IN superseded_path) = 0
        AND position('..' IN current_path) = 0
        AND position('//' IN current_path) = 0
        AND superseded_path <> current_path
        AND superseded_sha256 <> current_sha256)
);

INSERT INTO platform.tenant_migration_supersessions (
    owner,scope,logical_version,database_origin,
    superseded_path,superseded_sha256,current_path,current_sha256
) VALUES
(
    'system_resource_layout','tenant',20260730120104,'fresh',
    'core/internal/modules/system_resource_layout/migrations/tenant/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql',
    '561757a212e2205e14091db2f9c05ab94f97d5a8cc67438d16c44a2ed4d0a8a2',
    'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/fresh/20260730120104_repair_installation_layout_publications_v1.fresh.sql',
    '75ca56659160fe477fda1ddedec60389e378f410da3ef9e6b5dacfc885a98969'
),
(
    'system_resource_layout','tenant',20260730120104,'v1_adopted',
    'core/internal/modules/system_resource_layout/migrations/tenant/v1_adopted/20260730120104_repair_installation_layout_publications_v1.v1_adopted.sql',
    '561757a212e2205e14091db2f9c05ab94f97d5a8cc67438d16c44a2ed4d0a8a2',
    'core/internal/modules/system_resource_layout/migrations/tenant/supersessions/v1_adopted/20260730120104_repair_installation_layout_publications_v1.v1_adopted.sql',
    '75ca56659160fe477fda1ddedec60389e378f410da3ef9e6b5dacfc885a98969'
);

CREATE TRIGGER tenant_migration_supersessions_reject_mutation
BEFORE UPDATE OR DELETE ON platform.tenant_migration_supersessions
FOR EACH ROW
EXECUTE FUNCTION platform.platform_migrator_reject_mutation();

ALTER TABLE platform.tenant_migration_supersessions
    ENABLE ALWAYS TRIGGER tenant_migration_supersessions_reject_mutation;
