-- +goose Up
CREATE TABLE iam_one_time_tokens (
    id uuid NOT NULL,
    purpose text NOT NULL,
    user_id uuid,
    identifier_id uuid,
    identifier_kind text,
    identifier_revision bigint,
    invitation_id uuid,
    invitation_expires_at timestamptz,
    invitation_revision bigint,
    portal_binding_id uuid,
    portal_setup_revision bigint,
    selector bytea NOT NULL,
    hash_key_id text COLLATE "C" NOT NULL,
    secret_hmac bytea NOT NULL,
    binding_sha256 bytea NOT NULL,
    state text NOT NULL,
    attempt_count smallint NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    revoked_at timestamptz,
    expired_at timestamptz,
    purge_after timestamptz NOT NULL,
    CONSTRAINT iam_one_time_tokens_pkey PRIMARY KEY (id),
    CONSTRAINT iam_one_time_tokens_selector_key UNIQUE (selector),
    CONSTRAINT iam_one_time_tokens_secret_key
        UNIQUE (hash_key_id, secret_hmac),
    CONSTRAINT iam_one_time_tokens_user_fk
        FOREIGN KEY (user_id) REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_one_time_tokens_identifier_user_fk
        FOREIGN KEY (identifier_id, user_id, identifier_kind)
        REFERENCES iam_user_identifiers (id, user_id, kind)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_one_time_tokens_invitation_fk FOREIGN KEY (
        invitation_id, invitation_expires_at
    ) REFERENCES iam_invitations (id, expires_at)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_one_time_tokens_portal_binding_fk FOREIGN KEY (
        portal_binding_id, user_id
    ) REFERENCES iam_portal_subject_bindings (id, user_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_one_time_tokens_purpose_check CHECK (
        purpose IN (
            'password_reset_v1', 'email_verification_v1',
            'portal_setup_v1', 'invitation_accept_v1'
        )
    ),
    CONSTRAINT iam_one_time_tokens_subject_shape_check CHECK (
        (
            purpose = 'password_reset_v1' AND
            identifier_kind IN ('email_v1', 'portal_v1') AND
            identifier_revision IS NOT NULL AND identifier_revision > 0 AND
            user_id IS NOT NULL AND identifier_id IS NOT NULL AND
            invitation_id IS NULL AND invitation_expires_at IS NULL AND
            invitation_revision IS NULL AND portal_binding_id IS NULL AND
            portal_setup_revision IS NULL
        ) OR (
            purpose = 'email_verification_v1' AND
            identifier_kind = 'email_v1' AND user_id IS NOT NULL AND
            identifier_revision IS NOT NULL AND identifier_revision > 0 AND
            identifier_id IS NOT NULL AND invitation_id IS NULL AND
            invitation_expires_at IS NULL AND invitation_revision IS NULL AND
            portal_binding_id IS NULL AND
            portal_setup_revision IS NULL
        ) OR (
            purpose = 'portal_setup_v1' AND identifier_kind = 'portal_v1' AND
            identifier_revision IS NOT NULL AND identifier_revision > 0 AND
            user_id IS NOT NULL AND identifier_id IS NOT NULL AND
            invitation_id IS NULL AND invitation_expires_at IS NULL AND
            invitation_revision IS NULL AND
            portal_binding_id IS NOT NULL AND
            portal_setup_revision IS NOT NULL AND portal_setup_revision > 0
        ) OR (
            purpose = 'invitation_accept_v1' AND
            user_id IS NULL AND identifier_id IS NULL AND identifier_kind IS NULL AND
            identifier_revision IS NULL AND
            invitation_id IS NOT NULL AND invitation_expires_at IS NOT NULL AND
            invitation_revision IS NOT NULL AND invitation_revision > 0 AND
            portal_binding_id IS NULL AND portal_setup_revision IS NULL
        )
    ),
    CONSTRAINT iam_one_time_tokens_secret_check CHECK (
        pg_catalog.octet_length(selector) = 16 AND
        pg_catalog.char_length(hash_key_id) BETWEEN 1 AND 128 AND
        hash_key_id ~ '^[a-z0-9][a-z0-9._-]*$' AND
        pg_catalog.octet_length(secret_hmac) = 32 AND
        pg_catalog.octet_length(binding_sha256) = 32
    ),
    CONSTRAINT iam_one_time_tokens_state_check
        CHECK (state IN ('pending', 'consumed', 'revoked', 'expired')),
    CONSTRAINT iam_one_time_tokens_counter_check CHECK (
        attempt_count BETWEEN 0 AND 5 AND revision > 0
    ),
    CONSTRAINT iam_one_time_tokens_ttl_check CHECK (
        (
            purpose IN ('password_reset_v1', 'portal_setup_v1') AND
            expires_at = created_at + INTERVAL '1 hour'
        ) OR (
            purpose = 'email_verification_v1' AND
            expires_at = created_at + INTERVAL '24 hours'
        ) OR (
            purpose = 'invitation_accept_v1' AND
            expires_at = invitation_expires_at AND created_at < expires_at
        )
    ),
    CONSTRAINT iam_one_time_tokens_lifecycle_check CHECK (
        (
            state = 'pending' AND attempt_count < 5 AND
            consumed_at IS NULL AND revoked_at IS NULL AND expired_at IS NULL
        ) OR (
            state = 'consumed' AND consumed_at IS NOT NULL AND
            revoked_at IS NULL AND expired_at IS NULL
        ) OR (
            state = 'revoked' AND consumed_at IS NULL AND
            revoked_at IS NOT NULL AND expired_at IS NULL
        ) OR (
            state = 'expired' AND consumed_at IS NULL AND
            revoked_at IS NULL AND expired_at = expires_at
        )
    ),
    CONSTRAINT iam_one_time_tokens_time_check CHECK (
        updated_at >= created_at AND
        (consumed_at IS NULL OR
            (consumed_at >= created_at AND
                consumed_at <= expires_at AND consumed_at <= updated_at)) AND
        (revoked_at IS NULL OR
            (revoked_at >= created_at AND revoked_at <= updated_at)) AND
        (expired_at IS NULL OR expired_at <= updated_at) AND
        purge_after >= GREATEST(
            expires_at, COALESCE(consumed_at, expires_at),
            COALESCE(revoked_at, expires_at), COALESCE(expired_at, expires_at)
        ) + INTERVAL '24 hours'
    )
);

CREATE UNIQUE INDEX iam_one_time_tokens_one_pending_binding_uidx
    ON iam_one_time_tokens (purpose, binding_sha256)
    WHERE state = 'pending';

CREATE UNIQUE INDEX iam_one_time_tokens_one_pending_identifier_uidx
    ON iam_one_time_tokens (purpose, user_id, identifier_id)
    WHERE state = 'pending' AND identifier_id IS NOT NULL;

CREATE UNIQUE INDEX iam_one_time_tokens_one_pending_invitation_uidx
    ON iam_one_time_tokens (purpose, invitation_id)
    WHERE state = 'pending' AND invitation_id IS NOT NULL;

CREATE INDEX iam_one_time_tokens_retention_idx
    ON iam_one_time_tokens (purge_after)
    WHERE state <> 'pending';
