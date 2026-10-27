-- +goose Up
CREATE TABLE platform.tenancy_customer_provisioning_requests (
    operation_id uuid PRIMARY KEY REFERENCES platform.tenancy_onboarding_operations (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    control_tenant_id uuid NOT NULL REFERENCES platform.tenants (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    requested_by_subject_id uuid NOT NULL,
    customer_name text NOT NULL,
    tenant_slug text NOT NULL UNIQUE,
    first_admin_display_name text NOT NULL,
    first_admin_email text NOT NULL,
    request_sha256 text NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT tenancy_customer_provisioning_requests_shape_check CHECK (
        octet_length(customer_name) BETWEEN 1 AND 255
        AND tenant_slug ~ '^[a-z0-9](?:[a-z0-9-]{0,98}[a-z0-9])?$'
        AND octet_length(first_admin_display_name) BETWEEN 1 AND 255
        AND octet_length(first_admin_email) BETWEEN 3 AND 254
        AND first_admin_email = pg_catalog.lower(first_admin_email)
        AND request_sha256 ~ '^[0-9a-f]{64}$'
        AND pg_catalog.isfinite(created_at)
    )
);

CREATE TABLE platform.tenancy_customer_provisioning_outcomes (
    operation_id uuid PRIMARY KEY REFERENCES
        platform.tenancy_customer_provisioning_requests (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    setup_delivery_state text NOT NULL DEFAULT 'pending' CHECK (
        setup_delivery_state IN ('pending','delivered','outcome_unknown',
            'permanent_failure','accepted','reissued','revoked')
    ),
    setup_artifact_id uuid,
    failure_code text,
    failure_retryable boolean,
    correlation_id uuid,
    revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0),
    updated_at timestamptz NOT NULL,
    CONSTRAINT tenancy_customer_provisioning_outcomes_failure_check CHECK (
        (failure_code IS NULL AND failure_retryable IS NULL AND correlation_id IS NULL)
        OR (failure_code ~ '^[A-Z][A-Z0-9_]{2,63}$'
            AND failure_retryable IS NOT NULL AND correlation_id IS NOT NULL)
    )
);

CREATE INDEX tenancy_customer_provisioning_subject_idx
ON platform.tenancy_customer_provisioning_requests (
    requested_by_subject_id, created_at DESC, operation_id
);
