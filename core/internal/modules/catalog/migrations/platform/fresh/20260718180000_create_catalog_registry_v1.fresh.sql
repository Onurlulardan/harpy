-- +goose Up
CREATE TABLE platform.catalog_packages (
    package_id uuid PRIMARY KEY,
    slug text NOT NULL UNIQUE CHECK (slug ~ '^[a-z][a-z0-9-]{1,62}$'),
    created_at timestamptz NOT NULL,
    CHECK (date_trunc('microseconds', created_at) = created_at)
);

CREATE TABLE platform.catalog_package_versions (
    package_id uuid NOT NULL,
    version text NOT NULL CHECK (version ~ '^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'),
    manifest jsonb NOT NULL CHECK (jsonb_typeof(manifest) = 'object'),
    manifest_sha256 text NOT NULL CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'),
    signature bytea NOT NULL CHECK (octet_length(signature) = 64),
    key_id text NOT NULL CHECK (length(key_id) BETWEEN 1 AND 128),
    release_id uuid NOT NULL,
    registered_at timestamptz NOT NULL,
    PRIMARY KEY (package_id, version),
    CONSTRAINT catalog_package_versions_package_fk
        FOREIGN KEY (package_id) REFERENCES platform.catalog_packages(package_id)
        ON DELETE RESTRICT,
    CHECK (date_trunc('microseconds', registered_at) = registered_at)
);

-- +goose StatementBegin
CREATE FUNCTION platform.catalog_reject_registry_mutation_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY INVOKER
SET search_path = pg_catalog
AS $$
BEGIN
    RAISE EXCEPTION USING ERRCODE = '55000',
        MESSAGE = 'catalog registry is immutable';
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER catalog_packages_immutable
BEFORE UPDATE OR DELETE ON platform.catalog_packages
FOR EACH STATEMENT EXECUTE FUNCTION platform.catalog_reject_registry_mutation_v1();

CREATE TRIGGER catalog_package_versions_immutable
BEFORE UPDATE OR DELETE ON platform.catalog_package_versions
FOR EACH STATEMENT EXECUTE FUNCTION platform.catalog_reject_registry_mutation_v1();

GRANT SELECT ON TABLE platform.catalog_packages, platform.catalog_package_versions TO harpy_runtime;
