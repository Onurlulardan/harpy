-- +goose Up
ALTER TABLE platform.catalog_package_versions
    ADD CONSTRAINT catalog_package_versions_provenance_unique
    UNIQUE (package_id, version, release_id, registered_at);

ALTER TABLE platform.catalog_package_owner_payloads
    ADD CONSTRAINT catalog_package_owner_payloads_provenance_fk
    FOREIGN KEY (package_id, version, release_id, registered_at)
    REFERENCES platform.catalog_package_versions(
        package_id, version, release_id, registered_at
    )
    ON DELETE RESTRICT;
