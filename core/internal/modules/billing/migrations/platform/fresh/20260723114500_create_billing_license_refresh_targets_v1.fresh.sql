-- +goose Up
CREATE TABLE platform.billing_license_refresh_targets (
    license_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    licensee_organization_id uuid NOT NULL,
    delivery_kind text NOT NULL,
    observed_sequence bigint NOT NULL,
    observed_sha256 text NOT NULL,
    operation_id uuid NOT NULL UNIQUE,
    next_attempt_at timestamptz NOT NULL,
    lease_until timestamptz,
    revision bigint NOT NULL,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (license_id, environment_id),
    CONSTRAINT billing_license_refresh_target_head_fk
        FOREIGN KEY (license_id, environment_id)
        REFERENCES platform.billing_license_heads (license_id, environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_license_refresh_target_shape CHECK (
        license_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND environment_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND licensee_organization_id <>
            '00000000-0000-0000-0000-000000000000'::uuid
        AND delivery_kind IN ('connected', 'airgap', 'development')
        AND observed_sequence > 0
        AND observed_sha256 ~ '^[0-9a-f]{64}$'
        AND pg_catalog.uuid_extract_version(operation_id) = 7
        AND revision > 0
        AND pg_catalog.isfinite(next_attempt_at)
        AND pg_catalog.isfinite(updated_at)
        AND (lease_until IS NULL OR (
            pg_catalog.isfinite(lease_until) AND lease_until > updated_at
        ))
    )
);
