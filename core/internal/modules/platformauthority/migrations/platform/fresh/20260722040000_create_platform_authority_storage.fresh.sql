-- +goose Up
CREATE TABLE platform.platformauthority_bindings (
    binding_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL UNIQUE,
    deployment_mode text NOT NULL CHECK (deployment_mode IN ('onpremise', 'saas')),
    registry_kind text NOT NULL CHECK (registry_kind IN ('development', 'saas_provider')),
    license_binding_id uuid UNIQUE,
    finalization_receipt_sha256 text CHECK (finalization_receipt_sha256 ~ '^[0-9a-f]{64}$'),
    recovery_trust_bundle_sha256 text CHECK (recovery_trust_bundle_sha256 ~ '^[0-9a-f]{64}$'),
    signer_generation bigint NOT NULL CHECK (signer_generation > 0),
    signer_kid text NOT NULL CHECK (octet_length(signer_kid) BETWEEN 1 AND 512),
    signer_public_key_sha256 text NOT NULL CHECK (signer_public_key_sha256 ~ '^[0-9a-f]{64}$'),
    signer_status text NOT NULL DEFAULT 'active'
        CHECK (signer_status IN ('active', 'rekey_required')),
    proof_sha256 text CHECK (proof_sha256 ~ '^[0-9a-f]{64}$'),
    proof_bytes bytea,
    created_at timestamptz NOT NULL,
    CHECK (
        (deployment_mode = 'onpremise' AND registry_kind = 'development'
            AND license_binding_id IS NOT NULL AND finalization_receipt_sha256 IS NOT NULL
            AND recovery_trust_bundle_sha256 IS NOT NULL
            AND proof_sha256 IS NOT NULL AND proof_bytes IS NOT NULL)
        OR (deployment_mode = 'saas' AND registry_kind = 'saas_provider'
            AND license_binding_id IS NULL AND finalization_receipt_sha256 IS NULL
            AND recovery_trust_bundle_sha256 IS NULL
            AND proof_sha256 IS NULL AND proof_bytes IS NULL)
    )
);

CREATE TABLE platform.platformauthority_heads (
    binding_id uuid NOT NULL REFERENCES platform.platformauthority_bindings (binding_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    scope_kind text NOT NULL CHECK (scope_kind IN ('environment', 'tenant')),
    scope_id uuid NOT NULL,
    generation bigint NOT NULL CHECK (generation > 0),
    previous_head_sha256 text,
    state_sha256 text NOT NULL CHECK (state_sha256 ~ '^[0-9a-f]{64}$'),
    proposal_sha256 text NOT NULL CHECK (proposal_sha256 ~ '^[0-9a-f]{64}$'),
    ack_sha256 text NOT NULL UNIQUE CHECK (ack_sha256 ~ '^[0-9a-f]{64}$'),
    registry_sequence bigint NOT NULL CHECK (registry_sequence > 0),
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (binding_id, scope_kind, scope_id),
    UNIQUE (binding_id, registry_sequence),
    CHECK ((generation = 1) = (previous_head_sha256 IS NULL))
);

CREATE TABLE platform.platformauthority_operations (
    operation_id uuid PRIMARY KEY,
    binding_id uuid NOT NULL REFERENCES platform.platformauthority_bindings (binding_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    purpose text NOT NULL CHECK (purpose IN (
        'initial_environment_authority', 'initial_tenant_authority', 'owner_transfer',
        'protected_admin_change', 'recovery_anchor_change', 'proposal_signer_rotation',
        'compensate_uncommitted_ack'
    )),
    scope_kind text NOT NULL CHECK (scope_kind IN ('environment', 'tenant')),
    scope_id uuid NOT NULL,
    generation bigint NOT NULL CHECK (generation > 0),
    previous_head_sha256 text,
    state_sha256 text NOT NULL CHECK (state_sha256 ~ '^[0-9a-f]{64}$'),
    state_bytes bytea NOT NULL,
    proposal_sha256 text NOT NULL CHECK (proposal_sha256 ~ '^[0-9a-f]{64}$'),
    proposal_bytes bytea NOT NULL,
    compensates_ack_sha256 text CHECK (compensates_ack_sha256 ~ '^[0-9a-f]{64}$'),
    status text NOT NULL CHECK (status IN ('proposal_signed', 'ack_verified', 'completed')),
    ack_sha256 text CHECK (ack_sha256 ~ '^[0-9a-f]{64}$'),
    ack_bytes bytea,
    registry_sequence bigint CHECK (registry_sequence > 0),
    issued_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    ack_verified_at timestamptz,
    completed_at timestamptz,
    revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CHECK ((generation = 1) = (previous_head_sha256 IS NULL)),
    CHECK ((purpose = 'compensate_uncommitted_ack') = (compensates_ack_sha256 IS NOT NULL)),
    CHECK (expires_at > issued_at),
    CHECK ((status = 'proposal_signed' AND ack_sha256 IS NULL AND ack_bytes IS NULL
            AND registry_sequence IS NULL AND ack_verified_at IS NULL AND completed_at IS NULL)
        OR (status = 'ack_verified' AND ack_sha256 IS NOT NULL AND ack_bytes IS NOT NULL
            AND registry_sequence IS NOT NULL AND ack_verified_at IS NOT NULL AND completed_at IS NULL)
        OR (status = 'completed' AND ack_sha256 IS NOT NULL AND ack_bytes IS NOT NULL
            AND registry_sequence IS NOT NULL AND ack_verified_at IS NOT NULL AND completed_at IS NOT NULL))
);

CREATE UNIQUE INDEX platformauthority_operation_ack_uq
ON platform.platformauthority_operations (ack_sha256) WHERE ack_sha256 IS NOT NULL;

CREATE UNIQUE INDEX platformauthority_operation_sequence_uq
ON platform.platformauthority_operations (binding_id, registry_sequence)
WHERE registry_sequence IS NOT NULL;

CREATE TABLE platform.platformauthority_receipts (
    operation_id uuid PRIMARY KEY REFERENCES platform.platformauthority_operations (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    binding_id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    generation bigint NOT NULL CHECK (generation > 0),
    state_sha256 text NOT NULL CHECK (state_sha256 ~ '^[0-9a-f]{64}$'),
    proposal_sha256 text NOT NULL CHECK (proposal_sha256 ~ '^[0-9a-f]{64}$'),
    ack_sha256 text NOT NULL UNIQUE CHECK (ack_sha256 ~ '^[0-9a-f]{64}$'),
    registry_sequence bigint NOT NULL CHECK (registry_sequence > 0),
    completed_at timestamptz NOT NULL
);
