-- +goose Up
CREATE TABLE platform.billing_entitlement_registry_versions (
    registry_version bigint NOT NULL,
    authority_id uuid NOT NULL,
    source_contract text COLLATE "C" NOT NULL,
    canonical_sha256 bytea NOT NULL,
    definition_count integer NOT NULL,
    published_at timestamptz NOT NULL,
    CONSTRAINT billing_entitlement_registry_versions_pkey
        PRIMARY KEY (registry_version),
    CONSTRAINT billing_entitlement_registry_versions_authority_key
        UNIQUE (authority_id, registry_version),
    CONSTRAINT billing_entitlement_registry_versions_authority_fk
        FOREIGN KEY (authority_id)
        REFERENCES platform.billing_global_authorities (authority_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_entitlement_registry_versions_shape_check CHECK (
        registry_version > 0
        AND source_contract = 'ADR-0082'
        AND pg_catalog.octet_length(canonical_sha256) = 32
        AND definition_count > 0
    )
);

ALTER TABLE platform.billing_entitlement_definitions
    ADD CONSTRAINT billing_entitlement_definitions_registry_fk
    FOREIGN KEY (authority_id, registry_version)
    REFERENCES platform.billing_entitlement_registry_versions (
        authority_id, registry_version
    ) ON UPDATE NO ACTION ON DELETE RESTRICT;

