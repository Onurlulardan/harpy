-- +goose Up
CREATE TABLE platform.tenant_migration_requirements (
    migration_run_id uuid NOT NULL
        REFERENCES platform.tenant_migration_requirement_catalogs(migration_run_id)
        ON DELETE RESTRICT,
    ordinal integer NOT NULL CHECK (ordinal BETWEEN 1 AND 10000),
    owner text NOT NULL CHECK (owner ~ '^[a-z][a-z0-9_]{0,63}$'),
    logical_version bigint NOT NULL
        CHECK (logical_version BETWEEN 20000101000000 AND 99991231235959),
    database_origin text NOT NULL
        CHECK (database_origin IN ('fresh', 'v1_adopted')),
    selected_path text NOT NULL
        CHECK (selected_path ~ '^[a-z0-9][a-z0-9_./-]*\.sql$'
            AND pg_catalog.octet_length(selected_path) <= 512),
    selected_sha256 text NOT NULL
        CHECK (selected_sha256 ~ '^[0-9a-f]{64}$'),
    PRIMARY KEY (migration_run_id, ordinal),
    CONSTRAINT tenant_migration_requirements_identity_unique
        UNIQUE (migration_run_id, owner, logical_version),
    CHECK (position('..' IN selected_path) = 0
        AND position('//' IN selected_path) = 0)
);

CREATE TRIGGER tenant_migration_requirements_reject_mutation
BEFORE UPDATE OR DELETE ON platform.tenant_migration_requirements
FOR EACH ROW EXECUTE FUNCTION platform.platform_migrator_reject_mutation();

CREATE TABLE platform.tenant_migration_requirement_targets (
    migration_run_id uuid NOT NULL
        REFERENCES platform.tenant_migration_requirement_catalogs(migration_run_id)
        ON DELETE RESTRICT,
    ordinal integer NOT NULL CHECK (ordinal BETWEEN 1 AND 10000),
    tenant_id uuid NOT NULL,
    schema_name text NOT NULL CHECK (schema_name ~ '^[a-z][a-z0-9_]{0,62}$'),
    PRIMARY KEY (migration_run_id, ordinal),
    UNIQUE (migration_run_id, tenant_id),
    UNIQUE (migration_run_id, schema_name)
);

CREATE TRIGGER tenant_migration_requirement_targets_reject_mutation
BEFORE UPDATE OR DELETE ON platform.tenant_migration_requirement_targets
FOR EACH ROW EXECUTE FUNCTION platform.platform_migrator_reject_mutation();
