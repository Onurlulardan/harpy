-- +goose Up
CREATE TABLE platform.iam_recovery_anchors (
    anchor_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    owner_user_id uuid NOT NULL,
    generation bigint NOT NULL CHECK (generation > 0),
    credential_id bytea NOT NULL CHECK (pg_catalog.octet_length(credential_id) BETWEEN 1 AND 1024),
    public_key_cose bytea NOT NULL CHECK (pg_catalog.octet_length(public_key_cose) BETWEEN 1 AND 4096),
    aaguid uuid NOT NULL CHECK (aaguid <> '00000000-0000-0000-0000-000000000000'::uuid),
    attestation_format text COLLATE "C" NOT NULL CHECK (pg_catalog.octet_length(attestation_format) BETWEEN 1 AND 64),
    attestation_type text COLLATE "C" NOT NULL CHECK (pg_catalog.octet_length(attestation_type) BETWEEN 1 AND 64),
    rp_id text COLLATE "C" NOT NULL CHECK (pg_catalog.octet_length(rp_id) BETWEEN 1 AND 253),
    allowed_origins jsonb NOT NULL CHECK (pg_catalog.jsonb_typeof(allowed_origins) = 'array'),
    allowed_origins_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(allowed_origins_sha256) = 32),
    user_verification text NOT NULL DEFAULT 'required' CHECK (user_verification = 'required'),
    backup_eligible boolean NOT NULL DEFAULT false CHECK (backup_eligible = false),
    backup_state boolean NOT NULL DEFAULT false CHECK (backup_state = false),
    sign_count bigint NOT NULL CHECK (sign_count >= 0),
    counter_policy text NOT NULL DEFAULT 'adr0062' CHECK (counter_policy = 'adr0062'),
    state text NOT NULL CHECK (state IN ('pending','active','stale','deleted')),
    authority_generation bigint CHECK (authority_generation > 0),
    protected_state_sha256 bytea CHECK (
        protected_state_sha256 IS NULL OR pg_catalog.octet_length(protected_state_sha256) = 32
    ),
    authority_ack_sha256 bytea CHECK (
        authority_ack_sha256 IS NULL OR pg_catalog.octet_length(authority_ack_sha256) = 32
    ),
    revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    deleted_at timestamptz,
    CONSTRAINT iam_recovery_anchor_environment_fk FOREIGN KEY (environment_id)
        REFERENCES platform.tenancy_deployment_policy(environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_recovery_anchor_state_evidence_ck CHECK (
        (state = 'pending' AND authority_generation IS NULL AND
         protected_state_sha256 IS NULL AND authority_ack_sha256 IS NULL AND deleted_at IS NULL) OR
        (state IN ('active','stale') AND authority_generation IS NOT NULL AND
         protected_state_sha256 IS NOT NULL AND authority_ack_sha256 IS NOT NULL AND deleted_at IS NULL) OR
        (state = 'deleted' AND authority_generation IS NOT NULL AND
         protected_state_sha256 IS NOT NULL AND authority_ack_sha256 IS NOT NULL AND deleted_at IS NOT NULL)
    )
);

CREATE UNIQUE INDEX iam_recovery_anchor_credential_uq
ON platform.iam_recovery_anchors(environment_id,credential_id);
CREATE UNIQUE INDEX iam_recovery_anchor_generation_uq
ON platform.iam_recovery_anchors(environment_id,generation);

CREATE TABLE platform.iam_recovery_anchor_operations (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    actor_user_id uuid NOT NULL,
    kind text NOT NULL CHECK (kind IN ('anchor_register','anchor_delete')),
    anchor_id uuid NOT NULL,
    options_idempotency_key text COLLATE "C" NOT NULL CHECK (
        options_idempotency_key ~ '^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$'
    ),
    options_request_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(options_request_sha256) = 32),
    register_idempotency_key text COLLATE "C",
    register_request_sha256 bytea CHECK (
        register_request_sha256 IS NULL OR pg_catalog.octet_length(register_request_sha256) = 32
    ),
    challenge text COLLATE "C",
    challenge_expires_at timestamptz,
    rp_id text COLLATE "C",
    allowed_origins jsonb,
    allowed_origins_sha256 bytea CHECK (
        allowed_origins_sha256 IS NULL OR pg_catalog.octet_length(allowed_origins_sha256) = 32
    ),
    anchor_set_sha256 bytea CHECK (
        anchor_set_sha256 IS NULL OR pg_catalog.octet_length(anchor_set_sha256) = 32
    ),
    registry_binding_id uuid,
    authority_generation bigint CHECK (authority_generation > 0),
    protected_state_sha256 bytea CHECK (
        protected_state_sha256 IS NULL OR pg_catalog.octet_length(protected_state_sha256) = 32
    ),
    proposal_sha256 bytea CHECK (
        proposal_sha256 IS NULL OR pg_catalog.octet_length(proposal_sha256) = 32
    ),
    state text NOT NULL CHECK (state IN (
        'options_issued','authority_pending','proposal_signed','completed','failed'
    )),
    failure_code text COLLATE "C",
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    completed_at timestamptz,
    CONSTRAINT iam_recovery_anchor_operation_environment_fk FOREIGN KEY (environment_id)
        REFERENCES platform.tenancy_deployment_policy(environment_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_recovery_anchor_operation_shape_ck CHECK (
        (kind='anchor_register' AND challenge IS NOT NULL AND challenge_expires_at IS NOT NULL AND
         rp_id IS NOT NULL AND allowed_origins IS NOT NULL AND allowed_origins_sha256 IS NOT NULL) OR
        (kind='anchor_delete' AND challenge IS NULL AND challenge_expires_at IS NULL AND
         rp_id IS NULL AND allowed_origins IS NULL AND allowed_origins_sha256 IS NULL)
    ),
    CONSTRAINT iam_recovery_anchor_operation_state_ck CHECK (
        (state='options_issued' AND kind='anchor_register' AND anchor_set_sha256 IS NULL AND
         registry_binding_id IS NULL AND authority_generation IS NULL AND
         protected_state_sha256 IS NULL AND proposal_sha256 IS NULL AND completed_at IS NULL) OR
        (state='authority_pending' AND anchor_set_sha256 IS NOT NULL AND
         registry_binding_id IS NULL AND authority_generation IS NULL AND
         protected_state_sha256 IS NULL AND proposal_sha256 IS NULL AND completed_at IS NULL) OR
        (state='proposal_signed' AND anchor_set_sha256 IS NOT NULL AND
         registry_binding_id IS NOT NULL AND authority_generation IS NOT NULL AND
         protected_state_sha256 IS NOT NULL AND proposal_sha256 IS NOT NULL AND completed_at IS NULL) OR
        (state='completed' AND completed_at IS NOT NULL) OR state='failed'
    )
);

CREATE UNIQUE INDEX iam_recovery_anchor_operation_idempotency_uq
ON platform.iam_recovery_anchor_operations(environment_id,kind,options_idempotency_key);
CREATE UNIQUE INDEX iam_recovery_anchor_operation_open_uq
ON platform.iam_recovery_anchor_operations(environment_id)
WHERE state IN ('options_issued','authority_pending','proposal_signed');
