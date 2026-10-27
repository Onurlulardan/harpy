-- +goose Up
CREATE TABLE platform.iam_protected_authority_heads (
    environment_id uuid PRIMARY KEY,
    registry_binding_id uuid NOT NULL,
    scope_kind text NOT NULL CHECK (scope_kind = 'environment'),
    scope_id uuid NOT NULL UNIQUE,
    generation bigint NOT NULL CHECK (generation >= 0),
    state_sha256 bytea,
    ack_sha256 bytea,
    revision bigint NOT NULL CHECK (revision > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    CONSTRAINT iam_protected_heads_policy_fk FOREIGN KEY (environment_id)
        REFERENCES platform.iam_platform_owner_policies (environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_protected_heads_scope_ck CHECK (scope_id = environment_id),
    CONSTRAINT iam_protected_heads_digest_ck CHECK (
        (generation = 0 AND state_sha256 IS NULL AND ack_sha256 IS NULL) OR
        (generation > 0 AND pg_catalog.octet_length(state_sha256) = 32 AND
         pg_catalog.octet_length(ack_sha256) = 32)
    )
);

CREATE TABLE platform.iam_platform_owner_transfer_operations (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    request_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(request_sha256) = 32),
    actor_user_id uuid NOT NULL,
    successor_user_id uuid NOT NULL CHECK (successor_user_id <> actor_user_id),
    actor_expected_revision bigint NOT NULL CHECK (actor_expected_revision > 0),
    actor_expected_token_version bigint NOT NULL CHECK (actor_expected_token_version > 0),
    actor_expected_authorization_version bigint NOT NULL CHECK (actor_expected_authorization_version > 0),
    successor_expected_revision bigint NOT NULL CHECK (successor_expected_revision > 0),
    successor_expected_token_version bigint NOT NULL CHECK (successor_expected_token_version > 0),
    successor_expected_authorization_version bigint NOT NULL CHECK (successor_expected_authorization_version > 0),
    successor_subject_id uuid NOT NULL,
    successor_binding_id uuid NOT NULL,
    expected_owner_policy_revision bigint NOT NULL CHECK (expected_owner_policy_revision > 0),
    expected_platform_clock bigint NOT NULL CHECK (expected_platform_clock > 0),
    registry_binding_id uuid NOT NULL,
    scope_kind text NOT NULL CHECK (scope_kind = 'environment'),
    scope_id uuid NOT NULL,
    generation bigint NOT NULL CHECK (generation > 1),
    previous_head_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(previous_head_sha256) = 32),
    state_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(state_sha256) = 32),
    proposal_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(proposal_sha256) = 32),
    proposal_signer_kid text COLLATE "C" NOT NULL,
    proposal_signature_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(proposal_signature_sha256) = 32),
    proposal_expires_at timestamptz NOT NULL,
    status text NOT NULL CHECK (status IN ('prepared','ack_verified','completed','compensated')),
    registry_sequence bigint,
    ack_sha256 bytea,
    ack_registry_kid text COLLATE "C",
    ack_issued_at timestamptz,
    predecessor_token_version bigint,
    predecessor_authorization_version bigint,
    successor_token_version bigint,
    successor_authorization_version bigint,
    owner_policy_revision bigint,
    platform_clock bigint,
    revision bigint NOT NULL CHECK (revision > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    completed_at timestamptz,
    CONSTRAINT iam_owner_transfer_head_fk FOREIGN KEY (environment_id)
        REFERENCES platform.iam_protected_authority_heads (environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_owner_transfer_scope_ck CHECK (scope_id = environment_id),
    CONSTRAINT iam_owner_transfer_generation_uq UNIQUE (environment_id, generation),
    CONSTRAINT iam_owner_transfer_ack_uq UNIQUE (ack_sha256),
    CONSTRAINT iam_owner_transfer_ack_shape_ck CHECK (
        (status = 'prepared' AND registry_sequence IS NULL AND ack_sha256 IS NULL AND
         ack_registry_kid IS NULL AND ack_issued_at IS NULL) OR
        (status <> 'prepared' AND registry_sequence > 0 AND pg_catalog.octet_length(ack_sha256) = 32 AND
         ack_registry_kid IS NOT NULL AND ack_issued_at IS NOT NULL)
    ),
    CONSTRAINT iam_owner_transfer_result_shape_ck CHECK (
        (status IN ('prepared','ack_verified') AND predecessor_token_version IS NULL AND
         predecessor_authorization_version IS NULL AND successor_token_version IS NULL AND
         successor_authorization_version IS NULL AND owner_policy_revision IS NULL AND
         platform_clock IS NULL AND completed_at IS NULL) OR
        (status = 'completed' AND predecessor_token_version > 1 AND
         predecessor_authorization_version > 0 AND successor_token_version > 1 AND
         successor_authorization_version > 0 AND owner_policy_revision > 1 AND
         platform_clock > 1 AND completed_at IS NOT NULL) OR
        (status = 'compensated' AND predecessor_token_version IS NULL AND
         predecessor_authorization_version IS NULL AND successor_token_version IS NULL AND
         successor_authorization_version IS NULL AND owner_policy_revision IS NULL AND
         platform_clock IS NULL AND completed_at IS NOT NULL)
    )
);
