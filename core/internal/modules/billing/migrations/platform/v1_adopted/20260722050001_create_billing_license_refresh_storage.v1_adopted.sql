-- +goose Up
CREATE TABLE platform.billing_license_refresh_operations (
    operation_id uuid PRIMARY KEY,
    license_id uuid NOT NULL,
    licensee_organization_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    delivery_kind text NOT NULL,
    nonce bytea NOT NULL,
    component_versions jsonb NOT NULL,
    state text NOT NULL,
    current_attempt integer NOT NULL DEFAULT 0,
    lease_token uuid,
    lease_until timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    completed_at timestamptz,
    CONSTRAINT billing_license_refresh_operation_environment_fk
        FOREIGN KEY (environment_id)
        REFERENCES platform.environment_identity (environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_license_refresh_operation_shape_check CHECK (
        pg_catalog.uuid_extract_version(operation_id) = 7
        AND license_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND licensee_organization_id <>
            '00000000-0000-0000-0000-000000000000'::uuid
        AND delivery_kind IN ('connected', 'airgap', 'development')
        AND pg_catalog.octet_length(nonce) = 32
        AND pg_catalog.jsonb_typeof(component_versions) = 'array'
        AND pg_catalog.jsonb_array_length(component_versions) BETWEEN 1 AND 64
        AND state IN ('pending', 'running', 'succeeded', 'fenced')
        AND current_attempt >= 0
        AND created_at <= updated_at
    ),
    CONSTRAINT billing_license_refresh_operation_state_check CHECK (
        (state = 'running' AND lease_token IS NOT NULL
            AND lease_until IS NOT NULL AND completed_at IS NULL)
        OR (state = 'pending' AND lease_token IS NULL
            AND lease_until IS NULL AND completed_at IS NULL)
        OR (state IN ('succeeded', 'fenced') AND lease_token IS NULL
            AND lease_until IS NULL AND completed_at IS NOT NULL)
    )
);

CREATE TABLE platform.billing_license_refresh_attempts (
    operation_id uuid NOT NULL,
    attempt integer NOT NULL,
    lease_token uuid NOT NULL UNIQUE,
    state text NOT NULL,
    claimed_at timestamptz NOT NULL,
    lease_until timestamptz NOT NULL,
    finished_at timestamptz,
    failure_kind text,
    PRIMARY KEY (operation_id, attempt),
    CONSTRAINT billing_license_refresh_attempt_operation_fk
        FOREIGN KEY (operation_id)
        REFERENCES platform.billing_license_refresh_operations (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_license_refresh_attempt_shape_check CHECK (
        attempt > 0 AND claimed_at < lease_until
        AND state IN (
            'running', 'source_failed', 'verification_failed',
            'lease_expired', 'succeeded', 'noop',
            'sequence_conflict', 'sequence_rollback'
        )
    ),
    CONSTRAINT billing_license_refresh_attempt_state_check CHECK (
        (state = 'running' AND finished_at IS NULL AND failure_kind IS NULL)
        OR (state = 'source_failed' AND finished_at IS NOT NULL
            AND failure_kind = 'source_unavailable')
        OR (state = 'verification_failed' AND finished_at IS NOT NULL
            AND failure_kind = 'verification_failed')
        OR (state = 'lease_expired' AND finished_at IS NOT NULL
            AND failure_kind = 'lease_expired')
        OR (state = 'sequence_conflict' AND finished_at IS NOT NULL
            AND failure_kind = 'sequence_conflict')
        OR (state = 'sequence_rollback' AND finished_at IS NOT NULL
            AND failure_kind = 'sequence_rollback')
        OR (state IN ('succeeded', 'noop') AND finished_at IS NOT NULL
            AND failure_kind IS NULL)
    )
);

CREATE INDEX billing_license_refresh_pending_idx
ON platform.billing_license_refresh_operations (state, updated_at, operation_id);

