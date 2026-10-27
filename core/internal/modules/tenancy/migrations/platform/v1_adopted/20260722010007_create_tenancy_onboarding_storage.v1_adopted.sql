-- +goose Up
CREATE TABLE platform.tenancy_onboarding_operations (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    operation_kind text NOT NULL CHECK (operation_kind IN (
        'environment_initialize', 'saas_customer_tenant_provision'
    )),
    idempotency_key text NOT NULL CHECK (octet_length(idempotency_key) BETWEEN 1 AND 128),
    request_sha256 text NOT NULL CHECK (request_sha256 ~ '^[0-9a-f]{64}$'),
    manifest_sha256 text NOT NULL CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'),
    step_plan_sha256 text NOT NULL CHECK (step_plan_sha256 ~ '^[0-9a-f]{64}$'),
    deployment_mode text NOT NULL CHECK (deployment_mode IN ('onpremise', 'saas')),
    tenant_id uuid NOT NULL UNIQUE,
    schema_name name NOT NULL UNIQUE CHECK (schema_name::text ~ '^t_[a-z2-7]{26}$'),
    binding_purpose text NOT NULL CHECK (binding_purpose IN (
        'onpremise_primary', 'saas_control', 'saas_customer'
    )),
    first_user_id uuid NOT NULL UNIQUE,
    root_organization_id uuid NOT NULL UNIQUE,
    protected_admin_binding_id uuid NOT NULL UNIQUE,
    instance_owner_binding_id uuid UNIQUE,
    offering_version_id uuid NOT NULL,
    offering_revision bigint NOT NULL CHECK (offering_revision > 0),
    offering_sha256 text NOT NULL CHECK (offering_sha256 ~ '^[0-9a-f]{64}$'),
    offering_effective_at timestamptz NOT NULL,
    offering_source_sequence bigint,
    offering_source_sha256 text,
    recovery_signing_generation bigint,
    recovery_trust_bundle_sha256 text,
    registry_binding_id uuid NOT NULL,
    authority_scope_kind text NOT NULL CHECK (authority_scope_kind IN ('environment', 'tenant')),
    authority_scope_id uuid NOT NULL,
    authority_generation bigint NOT NULL CHECK (authority_generation >= 0),
    authority_previous_head_sha256 text,
    authority_state_sha256 text,
    authority_proposal_sha256 text,
    state text NOT NULL DEFAULT 'accepted' CHECK (state IN (
        'accepted', 'migrating', 'migration_ready', 'authority_pending',
        'authority_acknowledged', 'finalizing', 'succeeded', 'failed', 'fenced'
    )),
    revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT tenancy_onboarding_operations_policy_fk FOREIGN KEY (environment_id)
        REFERENCES platform.tenancy_deployment_policy (environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT tenancy_onboarding_operations_idempotency_uq
        UNIQUE (environment_id, operation_kind, idempotency_key),
    CONSTRAINT tenancy_onboarding_operations_profile_ck CHECK (
        (operation_kind = 'environment_initialize' AND deployment_mode = 'onpremise'
            AND binding_purpose = 'onpremise_primary' AND instance_owner_binding_id IS NOT NULL
            AND recovery_signing_generation > 0
            AND recovery_trust_bundle_sha256 ~ '^[0-9a-f]{64}$'
            AND offering_source_sequence > 0 AND offering_source_sha256 ~ '^[0-9a-f]{64}$'
            AND authority_scope_kind = 'environment' AND authority_scope_id = environment_id
            AND authority_generation = 0 AND authority_previous_head_sha256 IS NULL
            AND authority_state_sha256 IS NULL AND authority_proposal_sha256 IS NULL)
        OR (operation_kind = 'environment_initialize' AND deployment_mode = 'saas'
            AND binding_purpose = 'saas_control' AND instance_owner_binding_id IS NOT NULL
            AND recovery_signing_generation IS NULL AND recovery_trust_bundle_sha256 IS NULL
            AND offering_source_sequence IS NULL AND offering_source_sha256 IS NULL
            AND authority_scope_kind = 'environment' AND authority_scope_id = environment_id
            AND authority_generation = 1 AND authority_previous_head_sha256 IS NULL
            AND authority_state_sha256 ~ '^[0-9a-f]{64}$'
            AND authority_proposal_sha256 ~ '^[0-9a-f]{64}$')
        OR (operation_kind = 'saas_customer_tenant_provision' AND deployment_mode = 'saas'
            AND binding_purpose = 'saas_customer' AND instance_owner_binding_id IS NULL
            AND recovery_signing_generation IS NULL AND recovery_trust_bundle_sha256 IS NULL
            AND offering_source_sequence IS NULL AND offering_source_sha256 IS NULL
            AND authority_scope_kind = 'tenant' AND authority_scope_id = tenant_id
            AND authority_generation = 1 AND authority_previous_head_sha256 IS NULL
            AND authority_state_sha256 ~ '^[0-9a-f]{64}$'
            AND authority_proposal_sha256 ~ '^[0-9a-f]{64}$')
    )
);

CREATE UNIQUE INDEX tenancy_onboarding_environment_initialize_uq
ON platform.tenancy_onboarding_operations (environment_id)
WHERE operation_kind = 'environment_initialize';

CREATE TABLE platform.tenancy_onboarding_steps (
    operation_id uuid NOT NULL REFERENCES platform.tenancy_onboarding_operations (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    step_ordinal smallint NOT NULL CHECK (step_ordinal BETWEEN 1 AND 9),
    command_key text NOT NULL CHECK (octet_length(command_key) BETWEEN 3 AND 128),
    PRIMARY KEY (operation_id, step_ordinal),
    UNIQUE (operation_id, command_key)
);

CREATE TABLE platform.tenancy_onboarding_attempts (
    attempt_id uuid PRIMARY KEY,
    operation_id uuid NOT NULL REFERENCES platform.tenancy_onboarding_operations (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    attempt_number bigint NOT NULL CHECK (attempt_number > 0),
    outcome text NOT NULL CHECK (outcome IN ('active', 'expired', 'completed', 'failed', 'ambiguous')),
    started_at timestamptz NOT NULL,
    finished_at timestamptz,
    UNIQUE (operation_id, attempt_number),
    UNIQUE (attempt_id, operation_id),
    CHECK ((outcome = 'active' AND finished_at IS NULL) OR
        (outcome <> 'active' AND finished_at IS NOT NULL))
);

CREATE TABLE platform.tenancy_onboarding_leases (
    operation_id uuid PRIMARY KEY,
    attempt_id uuid NOT NULL,
    lease_owner text NOT NULL CHECK (octet_length(lease_owner) BETWEEN 1 AND 128),
    lease_token_sha256 text NOT NULL CHECK (lease_token_sha256 ~ '^[0-9a-f]{64}$'),
    lease_until timestamptz NOT NULL,
    heartbeat_at timestamptz NOT NULL,
    revision bigint NOT NULL CHECK (revision > 0),
    FOREIGN KEY (attempt_id, operation_id)
        REFERENCES platform.tenancy_onboarding_attempts (attempt_id, operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE TABLE platform.tenancy_onboarding_receipts (
    receipt_id uuid PRIMARY KEY,
    operation_id uuid NOT NULL REFERENCES platform.tenancy_onboarding_operations (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    attempt_id uuid NOT NULL,
    receipt_kind text NOT NULL CHECK (receipt_kind IN (
        'migration_claim_ready', 'migration_completed', 'activation_completed',
        'ambiguous_reconciled', 'failure'
    )),
    receipt_sha256 text NOT NULL CHECK (receipt_sha256 ~ '^[0-9a-f]{64}$'),
    evidence_sha256 text NOT NULL CHECK (evidence_sha256 ~ '^[0-9a-f]{64}$'),
    occurred_at timestamptz NOT NULL,
    FOREIGN KEY (attempt_id, operation_id)
        REFERENCES platform.tenancy_onboarding_attempts (attempt_id, operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    UNIQUE (operation_id, receipt_kind, receipt_sha256)
);

CREATE UNIQUE INDEX tenancy_onboarding_terminal_receipt_kind_uq
ON platform.tenancy_onboarding_receipts (operation_id, receipt_kind)
WHERE receipt_kind IN ('migration_completed', 'activation_completed', 'failure');

CREATE TABLE platform.tenancy_onboarding_authority_acks (
    operation_id uuid PRIMARY KEY REFERENCES platform.tenancy_onboarding_operations (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    registry_binding_id uuid NOT NULL,
    scope_kind text NOT NULL CHECK (scope_kind IN ('environment', 'tenant')),
    scope_id uuid NOT NULL,
    generation bigint NOT NULL CHECK (generation > 0),
    previous_head_sha256 text,
    state_sha256 text NOT NULL CHECK (state_sha256 ~ '^[0-9a-f]{64}$'),
    proposal_sha256 text NOT NULL CHECK (proposal_sha256 ~ '^[0-9a-f]{64}$'),
    ack_sha256 text NOT NULL UNIQUE CHECK (ack_sha256 ~ '^[0-9a-f]{64}$'),
    registry_sequence bigint NOT NULL CHECK (registry_sequence > 0),
    verified_at timestamptz NOT NULL
);
