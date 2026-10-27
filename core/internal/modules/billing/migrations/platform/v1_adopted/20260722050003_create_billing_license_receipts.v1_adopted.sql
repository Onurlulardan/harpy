-- +goose Up
CREATE TABLE platform.billing_license_verification_receipts (
    receipt_id uuid PRIMARY KEY DEFAULT pg_catalog.gen_random_uuid(),
    operation_id uuid NOT NULL,
    attempt integer NOT NULL,
    license_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    sequence bigint NOT NULL,
    payload_sha256 text COLLATE "C" NOT NULL,
    outcome text NOT NULL,
    verified_at timestamptz NOT NULL,
    CONSTRAINT billing_license_verification_attempt_fk
        FOREIGN KEY (operation_id, attempt)
        REFERENCES platform.billing_license_refresh_attempts (operation_id, attempt)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_license_verification_attempt_uq
        UNIQUE (operation_id, attempt),
    CONSTRAINT billing_license_verification_shape_check CHECK (
        sequence > 0 AND payload_sha256 ~ '^[0-9a-f]{64}$'
        AND outcome IN ('accepted', 'noop', 'sequence_conflict', 'sequence_rollback')
    )
);

CREATE TABLE platform.billing_license_heads (
    license_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    snapshot_id uuid NOT NULL,
    sequence bigint NOT NULL,
    payload_sha256 text COLLATE "C" NOT NULL,
    revision bigint NOT NULL,
    accepted_at timestamptz NOT NULL,
    PRIMARY KEY (license_id, environment_id),
    CONSTRAINT billing_license_head_snapshot_fk FOREIGN KEY (
        snapshot_id, license_id, environment_id, sequence, payload_sha256
    ) REFERENCES platform.billing_license_snapshots (
        snapshot_id, license_id, environment_id, sequence, payload_sha256
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_license_head_shape_check CHECK (
        sequence > 0 AND revision > 0
        AND payload_sha256 ~ '^[0-9a-f]{64}$'
    )
);

CREATE TABLE platform.billing_license_offering_publications (
    operation_id uuid PRIMARY KEY,
    authority_id uuid NOT NULL,
    license_id uuid NOT NULL,
    licensee_organization_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    delivery_kind text NOT NULL,
    issuer text COLLATE "C" NOT NULL,
    snapshot_id uuid NOT NULL,
    snapshot_sequence bigint NOT NULL,
    snapshot_sha256 text COLLATE "C" NOT NULL,
    offering_id uuid NOT NULL,
    offering_version_id uuid NOT NULL UNIQUE,
    revision bigint NOT NULL,
    offering_sha256 text COLLATE "C" NOT NULL,
    offering_document bytea NOT NULL,
    effective_at timestamptz NOT NULL,
    manifest_sha256 text COLLATE "C" NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT billing_license_offering_authority_fk
        FOREIGN KEY (authority_id)
        REFERENCES platform.billing_global_authorities (authority_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_license_offering_snapshot_fk FOREIGN KEY (
        snapshot_id, license_id, environment_id,
        snapshot_sequence, snapshot_sha256
    ) REFERENCES platform.billing_license_snapshots (
        snapshot_id, license_id, environment_id, sequence, payload_sha256
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_license_offering_root_fk
        FOREIGN KEY (authority_id, offering_id)
        REFERENCES platform.billing_offerings (authority_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_license_offering_version_fk
        FOREIGN KEY (authority_id, offering_version_id)
        REFERENCES platform.billing_offering_versions (authority_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_license_offering_snapshot_uq
        UNIQUE (license_id, environment_id, snapshot_sequence, snapshot_sha256),
    CONSTRAINT billing_license_offering_shape_check CHECK (
        pg_catalog.uuid_extract_version(operation_id) = 7
        AND pg_catalog.uuid_extract_version(offering_version_id) = 7
        AND delivery_kind IN ('connected', 'airgap', 'development')
        AND snapshot_sequence > 0 AND revision > 0
        AND snapshot_sha256 ~ '^[0-9a-f]{64}$'
        AND offering_sha256 ~ '^[0-9a-f]{64}$'
        AND manifest_sha256 ~ '^[0-9a-f]{64}$'
        AND pg_catalog.octet_length(offering_document) BETWEEN 2 AND 65536
        AND offering_sha256 = pg_catalog.encode(
            pg_catalog.sha256(offering_document), 'hex'
        )
    )
);

CREATE TABLE platform.billing_license_restore_receipts (
    operation_id uuid PRIMARY KEY,
    restore_receipt_id uuid NOT NULL UNIQUE,
    restore_receipt_sha256 text COLLATE "C" NOT NULL,
    license_binding_id uuid NOT NULL,
    license_id uuid NOT NULL,
    licensee_organization_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    previous_sequence bigint NOT NULL,
    previous_sha256 text COLLATE "C" NOT NULL,
    snapshot_id uuid NOT NULL,
    accepted_sequence bigint NOT NULL,
    accepted_sha256 text COLLATE "C" NOT NULL,
    expected_revision bigint NOT NULL,
    authority_revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT billing_license_restore_snapshot_fk FOREIGN KEY (
        snapshot_id, license_id, environment_id,
        accepted_sequence, accepted_sha256
    ) REFERENCES platform.billing_license_snapshots (
        snapshot_id, license_id, environment_id, sequence, payload_sha256
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_license_restore_accepted_uq
        UNIQUE (license_id, environment_id, accepted_sequence, accepted_sha256),
    CONSTRAINT billing_license_restore_shape_check CHECK (
        pg_catalog.uuid_extract_version(operation_id) = 7
        AND restore_receipt_id <>
            '00000000-0000-0000-0000-000000000000'::uuid
        AND license_binding_id <>
            '00000000-0000-0000-0000-000000000000'::uuid
        AND restore_receipt_sha256 ~ '^[0-9a-f]{64}$'
        AND previous_sequence > 0
        AND accepted_sequence > previous_sequence
        AND previous_sha256 ~ '^[0-9a-f]{64}$'
        AND accepted_sha256 ~ '^[0-9a-f]{64}$'
        AND expected_revision > 0
        AND authority_revision = expected_revision + 1
    )
);
