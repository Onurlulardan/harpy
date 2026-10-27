-- +goose Up
CREATE TABLE iam_refresh_tokens (
    id uuid NOT NULL,
    family_id uuid NOT NULL,
    family_absolute_expires_at timestamptz NOT NULL,
    parent_id uuid,
    parent_sequence bigint,
    selector bytea NOT NULL,
    hash_key_id text COLLATE "C" NOT NULL,
    secret_hmac bytea NOT NULL,
    sequence bigint NOT NULL,
    auth_time timestamptz NOT NULL,
    acr text NOT NULL,
    amr_password boolean NOT NULL,
    amr_webauthn boolean NOT NULL,
    amr_totp boolean NOT NULL,
    amr_recovery boolean NOT NULL,
    step_up_purpose_sha256 bytea,
    state text NOT NULL,
    revision bigint NOT NULL,
    issued_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    idle_expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    revoked_at timestamptz,
    expired_at timestamptz,
    retain_until timestamptz NOT NULL,
    CONSTRAINT iam_refresh_tokens_pkey PRIMARY KEY (id),
    CONSTRAINT iam_refresh_tokens_family_identity_key UNIQUE (id, family_id),
    CONSTRAINT iam_refresh_tokens_parent_identity_key
        UNIQUE (id, family_id, sequence),
    CONSTRAINT iam_refresh_tokens_selector_key UNIQUE (selector),
    CONSTRAINT iam_refresh_tokens_secret_key
        UNIQUE (hash_key_id, secret_hmac),
    CONSTRAINT iam_refresh_tokens_sequence_key UNIQUE (family_id, sequence),
    CONSTRAINT iam_refresh_tokens_family_fk FOREIGN KEY (
        family_id, family_absolute_expires_at
    ) REFERENCES iam_refresh_families (id, absolute_expires_at)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_refresh_tokens_parent_fk FOREIGN KEY (
        parent_id, family_id, parent_sequence
    ) REFERENCES iam_refresh_tokens (id, family_id, sequence)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_refresh_tokens_parent_check CHECK (
        (sequence = 1 AND parent_id IS NULL AND parent_sequence IS NULL) OR
        (sequence > 1 AND parent_id IS NOT NULL AND
            parent_sequence IS NOT NULL AND parent_id <> id AND
            sequence = parent_sequence + 1)
    ),
    CONSTRAINT iam_refresh_tokens_secret_check CHECK (
        pg_catalog.octet_length(selector) = 16 AND
        pg_catalog.char_length(hash_key_id) BETWEEN 1 AND 128 AND
        hash_key_id ~ '^[a-z0-9][a-z0-9._-]*$' AND
        pg_catalog.octet_length(secret_hmac) = 32
    ),
    CONSTRAINT iam_refresh_tokens_auth_context_check CHECK (
        acr IN ('login', 'mfa', 'step_up') AND
        (amr_password OR amr_webauthn OR amr_totp OR amr_recovery) AND
        (
            (acr IN ('login', 'mfa') AND step_up_purpose_sha256 IS NULL) OR
            (
                acr = 'step_up' AND NOT amr_recovery AND
                (amr_password OR amr_webauthn OR amr_totp) AND
                step_up_purpose_sha256 IS NOT NULL AND
                pg_catalog.octet_length(step_up_purpose_sha256) = 32
            )
        )
    ),
    CONSTRAINT iam_refresh_tokens_state_check
        CHECK (state IN ('active', 'consumed', 'revoked', 'expired')),
    CONSTRAINT iam_refresh_tokens_revision_check CHECK (
        sequence > 0 AND revision > 0
    ),
    CONSTRAINT iam_refresh_tokens_lifecycle_check CHECK (
        (
            state = 'active' AND consumed_at IS NULL AND
            revoked_at IS NULL AND expired_at IS NULL
        ) OR (
            state = 'consumed' AND consumed_at IS NOT NULL AND
            revoked_at IS NULL AND expired_at IS NULL
        ) OR (
            state = 'revoked' AND consumed_at IS NULL AND
            revoked_at IS NOT NULL AND expired_at IS NULL
        ) OR (
            state = 'expired' AND consumed_at IS NULL AND
            revoked_at IS NULL AND expired_at = idle_expires_at
        )
    ),
    CONSTRAINT iam_refresh_tokens_time_check CHECK (
        updated_at >= issued_at AND auth_time <= issued_at AND
        family_absolute_expires_at > issued_at AND
        idle_expires_at > issued_at AND
        idle_expires_at <= issued_at + INTERVAL '30 days' AND
        idle_expires_at <= family_absolute_expires_at AND
        (consumed_at IS NULL OR
            (consumed_at >= issued_at AND consumed_at <= updated_at)) AND
        (revoked_at IS NULL OR
            (revoked_at >= issued_at AND revoked_at <= updated_at)) AND
        (expired_at IS NULL OR expired_at <= updated_at) AND
        retain_until >= family_absolute_expires_at + INTERVAL '24 hours'
    )
);

CREATE UNIQUE INDEX iam_refresh_tokens_one_successor_uidx
    ON iam_refresh_tokens (parent_id)
    WHERE parent_id IS NOT NULL;

CREATE UNIQUE INDEX iam_refresh_tokens_one_active_family_uidx
    ON iam_refresh_tokens (family_id)
    WHERE state = 'active';

CREATE INDEX iam_refresh_tokens_retention_idx
    ON iam_refresh_tokens (retain_until)
    WHERE state <> 'active';
