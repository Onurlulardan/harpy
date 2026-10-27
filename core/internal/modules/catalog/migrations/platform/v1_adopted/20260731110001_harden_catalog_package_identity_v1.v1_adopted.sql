-- +goose Up
ALTER TABLE platform.catalog_packages
    ADD CONSTRAINT catalog_packages_package_id_nonzero CHECK (
        package_id IS NOT NULL AND
        package_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ) NOT VALID;
ALTER TABLE platform.catalog_packages
    VALIDATE CONSTRAINT catalog_packages_package_id_nonzero;

ALTER TABLE platform.catalog_package_versions
    ADD CONSTRAINT catalog_package_versions_package_id_nonzero CHECK (
        package_id IS NOT NULL AND
        package_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ) NOT VALID;
ALTER TABLE platform.catalog_package_versions
    VALIDATE CONSTRAINT catalog_package_versions_package_id_nonzero;

ALTER TABLE platform.catalog_package_owner_payloads
    ADD CONSTRAINT catalog_owner_payloads_package_id_nonzero CHECK (
        package_id IS NOT NULL AND
        package_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ) NOT VALID;
ALTER TABLE platform.catalog_package_owner_payloads
    VALIDATE CONSTRAINT catalog_owner_payloads_package_id_nonzero;
