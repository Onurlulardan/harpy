-- +goose Up
CREATE TABLE platform.tenant_migration_requirement_catalogs (
    migration_run_id uuid PRIMARY KEY
        REFERENCES platform.release_generations(migration_run_id) ON DELETE RESTRICT,
    contract_version text NOT NULL
        CHECK (contract_version = 'tenant-migration-requirement-catalog.v2'),
    release_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    database_origin text NOT NULL
        CHECK (database_origin IN ('fresh', 'v1_adopted')),
    migration_manifest_sha256 text NOT NULL
        CHECK (migration_manifest_sha256 ~ '^[0-9a-f]{64}$'),
    catalog_sha256 text NOT NULL UNIQUE
        CHECK (catalog_sha256 ~ '^[0-9a-f]{64}$'),
    requirement_count integer NOT NULL
        CHECK (requirement_count BETWEEN 1 AND 10000),
    target_count integer NOT NULL
        CHECK (target_count BETWEEN 0 AND 10000),
    canonical_document bytea NOT NULL
        CHECK (pg_catalog.octet_length(canonical_document) BETWEEN 2 AND 16777216),
    recorded_at timestamptz NOT NULL DEFAULT pg_catalog.statement_timestamp(),
    CONSTRAINT tenant_requirement_catalogs_release_unique
        UNIQUE (environment_id, release_id),
    CONSTRAINT tenant_requirement_catalogs_document_digest_check CHECK (
        catalog_sha256 = pg_catalog.encode(
            pg_catalog.sha256(canonical_document), 'hex'
        )
    )
);

CREATE TRIGGER tenant_migration_requirement_catalogs_reject_mutation
BEFORE UPDATE OR DELETE ON platform.tenant_migration_requirement_catalogs
FOR EACH ROW EXECUTE FUNCTION platform.platform_migrator_reject_mutation();
