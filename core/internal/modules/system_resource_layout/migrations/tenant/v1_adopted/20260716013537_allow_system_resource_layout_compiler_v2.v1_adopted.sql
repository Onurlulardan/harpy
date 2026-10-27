-- +goose Up
ALTER TABLE system_resource_layout_publications
    DROP CONSTRAINT system_resource_layout_publications_catalog_check;

ALTER TABLE system_resource_layout_publications
    ADD CONSTRAINT system_resource_layout_publications_catalog_check CHECK (
        catalog_contract IS NULL OR (
            compiler_version IN (
                'system-resource-presentation-compiler.v1',
                'system-resource-presentation-compiler.v2'
            )
            AND catalog_contract = 'forms-presentation-field-catalog.v1'
            AND catalog_json::jsonb ->> 'contract' = catalog_contract
            AND pg_catalog.octet_length(catalog_sha256) = 32
            AND catalog_sha256 = pg_catalog.sha256(
                pg_catalog.convert_to(catalog_json, 'UTF8')
            )
        )
    );
