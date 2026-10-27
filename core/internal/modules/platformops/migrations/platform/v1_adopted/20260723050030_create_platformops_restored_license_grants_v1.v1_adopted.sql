-- +goose Up
CREATE TABLE platform.platformops_restored_license_grants (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    restore_receipt_id uuid NOT NULL UNIQUE,
    restore_receipt_sha256 text NOT NULL,
    license_binding_id uuid NOT NULL UNIQUE,
    license_id uuid NOT NULL,
    licensee_organization_id uuid NOT NULL,
    delivery_kind text NOT NULL,
    issuer text NOT NULL,
    previous_sequence bigint NOT NULL,
    previous_sha256 text NOT NULL,
    next_sequence bigint NOT NULL,
    next_sha256 text NOT NULL,
    manifest_sha256 text NOT NULL,
    authority_revision bigint NOT NULL,
    lease_until timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT platformops_restored_license_operation_fk
        FOREIGN KEY (operation_id)
        REFERENCES platform.platformops_restore_operations (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT platformops_restored_license_shape CHECK (
        pg_catalog.uuid_extract_version(operation_id) = 7
        AND pg_catalog.uuid_extract_version(license_binding_id) = 7
        AND environment_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND restore_receipt_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND license_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND licensee_organization_id <>
            '00000000-0000-0000-0000-000000000000'::uuid
        AND restore_receipt_sha256 ~ '^[0-9a-f]{64}$'
        AND delivery_kind IN ('connected','airgap','development')
        AND pg_catalog.octet_length(issuer) BETWEEN 1 AND 256
        AND issuer = pg_catalog.btrim(issuer) AND issuer !~ '[[:cntrl:]]'
        AND previous_sequence > 0 AND next_sequence > previous_sequence
        AND previous_sha256 ~ '^[0-9a-f]{64}$'
        AND next_sha256 ~ '^[0-9a-f]{64}$'
        AND manifest_sha256 ~ '^[0-9a-f]{64}$'
        AND authority_revision > 0
        AND pg_catalog.isfinite(lease_until)
        AND created_at <= updated_at AND updated_at < lease_until
    )
);
