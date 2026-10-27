-- +goose Up
CREATE TABLE platform.iam_platform_owner_transfer_compensations (
    operation_id uuid PRIMARY KEY,
    original_operation_id uuid NOT NULL UNIQUE,
    environment_id uuid NOT NULL,
    request_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(request_sha256) = 32),
    actor_user_id uuid NOT NULL,
    expected_platform_clock bigint NOT NULL CHECK (expected_platform_clock > 0),
    registry_binding_id uuid NOT NULL,
    scope_kind text NOT NULL CHECK (scope_kind = 'environment'),
    scope_id uuid NOT NULL,
    generation bigint NOT NULL CHECK (generation > 2),
    previous_head_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(previous_head_sha256) = 32),
    state_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(state_sha256) = 32),
    proposal_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(proposal_sha256) = 32),
    proposal_signer_kid text COLLATE "C" NOT NULL,
    proposal_signature_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(proposal_signature_sha256) = 32),
    proposal_expires_at timestamptz NOT NULL,
    compensates_ack_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(compensates_ack_sha256) = 32),
    status text NOT NULL CHECK (status IN ('prepared','completed')),
    registry_sequence bigint,
    ack_sha256 bytea,
    ack_registry_kid text COLLATE "C",
    ack_issued_at timestamptz,
    revision bigint NOT NULL CHECK (revision > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    completed_at timestamptz,
    CONSTRAINT iam_owner_compensation_original_fk FOREIGN KEY (original_operation_id)
        REFERENCES platform.iam_platform_owner_transfer_operations (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_owner_compensation_head_fk FOREIGN KEY (environment_id)
        REFERENCES platform.iam_protected_authority_heads (environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_owner_compensation_scope_ck CHECK (scope_id = environment_id),
    CONSTRAINT iam_owner_compensation_generation_uq UNIQUE (environment_id, generation),
    CONSTRAINT iam_owner_compensation_ack_uq UNIQUE (ack_sha256),
    CONSTRAINT iam_owner_compensation_result_ck CHECK (
        (status = 'prepared' AND registry_sequence IS NULL AND ack_sha256 IS NULL AND
         ack_registry_kid IS NULL AND ack_issued_at IS NULL AND completed_at IS NULL) OR
        (status = 'completed' AND registry_sequence > 0 AND
         pg_catalog.octet_length(ack_sha256) = 32 AND ack_registry_kid IS NOT NULL AND
         ack_issued_at IS NOT NULL AND completed_at IS NOT NULL)
    )
);
