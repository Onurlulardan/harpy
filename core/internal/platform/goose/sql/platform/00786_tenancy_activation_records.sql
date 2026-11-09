-- +goose Up
CREATE TABLE IF NOT EXISTS platform.tenancy_activation_facts (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    owner_user_id uuid NOT NULL,
    schema_name text NOT NULL,
    kind text NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    binding_purpose text NOT NULL,
    manifest_sha256 text NOT NULL,
    activated_at timestamptz NOT NULL,
    CONSTRAINT tenancy_activation_facts_schema_name_check
        CHECK (schema_name ~ '^tenant_[0-9a-f]{32}$')
);

CREATE TABLE IF NOT EXISTS platform.tenancy_onboarding_attempts (
    attempt_id uuid PRIMARY KEY,
    operation_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    attempt_number bigint NOT NULL,
    lease_owner text NOT NULL,
    lease_token_sha256 text NOT NULL,
    created_at timestamptz NOT NULL,
    UNIQUE (operation_id, attempt_number)
);

CREATE TABLE IF NOT EXISTS platform.tenancy_onboarding_leases (
    operation_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    attempt_id uuid NOT NULL,
    lease_owner text NOT NULL,
    lease_token_sha256 text NOT NULL,
    lease_until timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS platform.tenancy_onboarding_receipts (
    operation_id uuid NOT NULL,
    receipt_kind text NOT NULL,
    receipt_id uuid NOT NULL,
    receipt_sha256 text NOT NULL,
    evidence_sha256 text NOT NULL,
    created_at timestamptz NOT NULL,
    PRIMARY KEY (operation_id, receipt_kind)
);

CREATE TABLE IF NOT EXISTS platform.tenancy_onboarding_authority_acks (
    operation_id uuid PRIMARY KEY,
    registry_binding_id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    generation bigint NOT NULL,
    previous_head_sha256 text NOT NULL,
    state_sha256 text NOT NULL,
    proposal_sha256 text NOT NULL,
    ack_sha256 text NOT NULL,
    registry_sequence bigint NOT NULL,
    verified_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS platform.tenancy_activation_step_capabilities (
    operation_id uuid NOT NULL,
    ordinal integer NOT NULL,
    command_name text NOT NULL,
    tenant_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    profile_kind text NOT NULL,
    manifest_sha256 text NOT NULL,
    schema_name text NOT NULL,
    session_user_name name NOT NULL,
    backend_pid integer NOT NULL,
    transaction_id xid8 NOT NULL,
    created_at timestamptz NOT NULL,
    PRIMARY KEY (operation_id, command_name),
    UNIQUE (operation_id, ordinal)
);

REVOKE ALL ON TABLE platform.tenancy_activation_facts FROM PUBLIC, harpy_api, harpy_worker;
REVOKE ALL ON TABLE platform.tenancy_onboarding_receipts FROM PUBLIC, harpy_api, harpy_worker;
REVOKE ALL ON TABLE platform.tenancy_onboarding_authority_acks FROM PUBLIC, harpy_api, harpy_worker;
REVOKE ALL ON TABLE platform.tenancy_activation_step_capabilities FROM PUBLIC, harpy_api, harpy_worker;
REVOKE ALL ON TABLE platform.tenancy_onboarding_attempts FROM PUBLIC;
REVOKE ALL ON TABLE platform.tenancy_onboarding_leases FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE ON TABLE platform.tenancy_onboarding_attempts TO harpy_api, harpy_worker;
GRANT SELECT, INSERT, UPDATE ON TABLE platform.tenancy_onboarding_leases TO harpy_api, harpy_worker;

-- +goose Down
DROP TABLE IF EXISTS platform.tenancy_activation_step_capabilities;
DROP TABLE IF EXISTS platform.tenancy_onboarding_authority_acks;
DROP TABLE IF EXISTS platform.tenancy_onboarding_receipts;
DROP TABLE IF EXISTS platform.tenancy_onboarding_leases;
DROP TABLE IF EXISTS platform.tenancy_onboarding_attempts;
DROP TABLE IF EXISTS platform.tenancy_activation_facts;
