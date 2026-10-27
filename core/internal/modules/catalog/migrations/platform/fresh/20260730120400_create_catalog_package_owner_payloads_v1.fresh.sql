-- +goose Up
CREATE TABLE platform.catalog_package_owner_payloads (
    package_id uuid NOT NULL,
    version text NOT NULL,
    component_id uuid NOT NULL,
    owner text NOT NULL CHECK (owner ~ '^[a-z][a-z0-9_]{0,62}$'),
    contract text NOT NULL CHECK (contract ~ '^[a-z][a-z0-9.-]{2,63}$'),
    content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    document jsonb NOT NULL CHECK (jsonb_typeof(document) = 'object'),
    release_id uuid NOT NULL,
    registered_at timestamptz NOT NULL,
    PRIMARY KEY (package_id, version, component_id),
    CONSTRAINT catalog_package_owner_payloads_version_fk
        FOREIGN KEY (package_id, version)
        REFERENCES platform.catalog_package_versions(package_id, version)
        ON DELETE RESTRICT,
    CHECK (date_trunc('microseconds', registered_at) = registered_at)
);

CREATE TRIGGER catalog_package_owner_payloads_immutable
BEFORE UPDATE OR DELETE ON platform.catalog_package_owner_payloads
FOR EACH STATEMENT EXECUTE FUNCTION platform.catalog_reject_registry_mutation_v1();

GRANT SELECT ON TABLE platform.catalog_package_owner_payloads TO forgecrud_runtime;
