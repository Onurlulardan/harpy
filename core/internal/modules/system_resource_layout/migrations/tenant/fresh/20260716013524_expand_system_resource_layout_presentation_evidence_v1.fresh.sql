-- +goose Up
ALTER TABLE system_resource_layout_publications
    ADD COLUMN source_definition_sha256 bytea,
    ADD COLUMN compiler_version text,
    ADD COLUMN catalog_contract text,
    ADD COLUMN catalog_json text,
    ADD COLUMN catalog_sha256 bytea,
    ADD COLUMN presentation_contract text,
    ADD COLUMN presentation_json text,
    ADD COLUMN presentation_sha256 bytea,
    ADD CONSTRAINT system_resource_layout_publications_evidence_complete_check
        CHECK (
            (
                source_definition_sha256 IS NULL
                AND compiler_version IS NULL
                AND catalog_contract IS NULL
                AND catalog_json IS NULL
                AND catalog_sha256 IS NULL
                AND presentation_contract IS NULL
                AND presentation_json IS NULL
                AND presentation_sha256 IS NULL
            ) OR (
                source_definition_sha256 IS NOT NULL
                AND compiler_version IS NOT NULL
                AND catalog_contract IS NOT NULL
                AND catalog_json IS NOT NULL
                AND catalog_sha256 IS NOT NULL
                AND presentation_contract IS NOT NULL
                AND presentation_json IS NOT NULL
                AND presentation_sha256 IS NOT NULL
            )
        ),
    ADD CONSTRAINT system_resource_layout_publications_source_digest_check
        CHECK (
            source_definition_sha256 IS NULL
            OR pg_catalog.octet_length(source_definition_sha256) = 32
        ),
    ADD CONSTRAINT system_resource_layout_publications_catalog_check CHECK (
        catalog_contract IS NULL OR (
            compiler_version = 'system-resource-presentation-compiler.v1'
            AND catalog_contract = 'forms-presentation-field-catalog.v1'
            AND catalog_json::jsonb ->> 'contract' = catalog_contract
            AND pg_catalog.octet_length(catalog_sha256) = 32
            AND catalog_sha256 = pg_catalog.sha256(
                pg_catalog.convert_to(catalog_json, 'UTF8')
            )
        )
    ),
    ADD CONSTRAINT system_resource_layout_publications_presentation_check CHECK (
        presentation_contract IS NULL OR (
            presentation_contract = 'forms-presentation.v1'
            AND presentation_json::jsonb ->> 'registry_contract' =
                'forms-presentation-registry.v1'
            AND pg_catalog.octet_length(presentation_sha256) = 32
            AND presentation_sha256 = pg_catalog.sha256(
                pg_catalog.convert_to(presentation_json, 'UTF8')
            )
        )
    );
