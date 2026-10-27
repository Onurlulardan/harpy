-- +goose Up
CREATE TABLE platform.tenancy_installation_credential_deliveries (
    operation_id uuid PRIMARY KEY REFERENCES platform.tenancy_onboarding_operations(operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    tenant_id uuid NOT NULL UNIQUE,
    first_owner_user_id uuid NOT NULL UNIQUE,
    installation_manifest_sha256 text NOT NULL CHECK (
        installation_manifest_sha256 ~ '^[0-9a-f]{64}$'),
    delivery_manifest_sha256 text NOT NULL CHECK (
        delivery_manifest_sha256 ~ '^[0-9a-f]{64}$'),
    current_attempt_id uuid NOT NULL UNIQUE,
    current_credential_id uuid NOT NULL UNIQUE,
    current_credential_version bigint NOT NULL CHECK (current_credential_version > 0),
    state text NOT NULL CHECK (state IN (
        'ready_to_display','claimed','acknowledged','outcome_unknown',
        'secret_deleted','completed')),
    claim_adapter_id text,
    claim_token_sha256 text,
    claim_lease_until timestamptz,
    claimed_at timestamptz,
    acknowledged_at timestamptz,
    secret_deleted_at timestamptz,
    completed_at timestamptz,
    recovery_attempt_id uuid,
    recovery_credential_id uuid,
    recovery_credential_version bigint,
    recovery_token_sha256 text,
    recovery_lease_until timestamptz,
    recovery_cause text,
    revision bigint NOT NULL CHECK (revision > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CHECK ((claim_adapter_id IS NULL AND claim_token_sha256 IS NULL AND
            claim_lease_until IS NULL AND claimed_at IS NULL) OR
           (octet_length(claim_adapter_id) BETWEEN 1 AND 128 AND
            claim_token_sha256 ~ '^[0-9a-f]{64}$' AND
            claim_lease_until IS NOT NULL AND claimed_at IS NOT NULL)),
    CHECK ((recovery_attempt_id IS NULL AND recovery_credential_id IS NULL AND
            recovery_credential_version IS NULL AND recovery_token_sha256 IS NULL AND
            recovery_lease_until IS NULL AND recovery_cause IS NULL) OR
           (state = 'outcome_unknown' AND recovery_attempt_id IS NOT NULL AND
            recovery_credential_id IS NOT NULL AND
            recovery_credential_version = current_credential_version + 1 AND
            recovery_token_sha256 ~ '^[0-9a-f]{64}$' AND
            recovery_lease_until IS NOT NULL AND recovery_cause IN (
                'claim_lease_expired','outcome_unknown'))),
    CHECK ((state = 'ready_to_display' AND claim_adapter_id IS NULL) OR
           (state IN ('claimed','outcome_unknown') AND claim_adapter_id IS NOT NULL) OR
           (state IN ('acknowledged','secret_deleted','completed') AND
                claim_adapter_id IS NOT NULL AND acknowledged_at IS NOT NULL)),
    CHECK ((state IN ('secret_deleted','completed')) = (secret_deleted_at IS NOT NULL)),
    CHECK ((state = 'completed') = (completed_at IS NOT NULL))
);

CREATE TABLE platform.tenancy_installation_credential_delivery_attempts (
    attempt_id uuid PRIMARY KEY,
    operation_id uuid NOT NULL REFERENCES
        platform.tenancy_installation_credential_deliveries(operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    credential_id uuid NOT NULL UNIQUE,
    credential_version bigint NOT NULL CHECK (credential_version > 0),
    state text NOT NULL CHECK (state IN (
        'ready_to_display','claimed','acknowledged','outcome_unknown',
        'superseded','completed')),
    cause text CHECK (cause IN ('claim_lease_expired','outcome_unknown')),
    superseded_by_attempt_id uuid,
    created_at timestamptz NOT NULL,
    claimed_at timestamptz,
    acknowledged_at timestamptz,
    superseded_at timestamptz,
    completed_at timestamptz,
    UNIQUE(operation_id, credential_version),
    CHECK ((state IN ('outcome_unknown','superseded')) = (cause IS NOT NULL)),
    CHECK ((state = 'superseded') =
        (superseded_by_attempt_id IS NOT NULL AND superseded_at IS NOT NULL)),
    CHECK ((state = 'completed') = (completed_at IS NOT NULL))
);

ALTER TABLE platform.tenancy_installation_credential_deliveries
ADD CONSTRAINT tenancy_installation_delivery_attempt_fk
FOREIGN KEY (current_attempt_id) REFERENCES
    platform.tenancy_installation_credential_delivery_attempts(attempt_id)
    DEFERRABLE INITIALLY DEFERRED;
