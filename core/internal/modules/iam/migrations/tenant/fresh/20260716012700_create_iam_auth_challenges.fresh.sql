-- +goose Up
CREATE TABLE iam_auth_challenges (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    session_id uuid,
    authenticator_id uuid,
    authenticator_kind text,
    purpose text NOT NULL,
    proof_kind text NOT NULL,
    hash_profile text NOT NULL,
    hmac_key_id text COLLATE "C" NOT NULL,
    challenge_hmac bytea NOT NULL,
    context_sha256 bytea NOT NULL,
    step_up_purpose_sha256 bytea,
    origin_policy_version bigint,
    action_policy_version bigint,
    rp_id_sha256 bytea,
    origin_sha256 bytea,
    user_verification_required boolean,
    state text NOT NULL,
    attempt_count smallint NOT NULL,
    revision bigint NOT NULL,
    issued_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    terminal_at timestamptz,
    purge_after timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT iam_auth_challenges_pkey PRIMARY KEY (id),
    CONSTRAINT iam_auth_challenges_secret_key
        UNIQUE (hmac_key_id, challenge_hmac),
    CONSTRAINT iam_auth_challenges_user_fk
        FOREIGN KEY (user_id) REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_auth_challenges_session_fk
        FOREIGN KEY (session_id, user_id) REFERENCES iam_sessions (id, user_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_auth_challenges_authenticator_fk
        FOREIGN KEY (authenticator_id, user_id, authenticator_kind)
        REFERENCES iam_authenticators (id, user_id, kind)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_auth_challenges_purpose_check CHECK (
        purpose IN (
            'webauthn_register_v1', 'webauthn_authenticate_v1',
            'totp_enroll_v1', 'reauth_v1'
        )
    ),
    CONSTRAINT iam_auth_challenges_proof_check
        CHECK (proof_kind IN ('password_v1', 'webauthn_v1', 'totp_v1')),
    CONSTRAINT iam_auth_challenges_purpose_shape_check CHECK (
        (
            purpose = 'webauthn_register_v1' AND
            proof_kind = 'webauthn_v1' AND session_id IS NOT NULL AND
            authenticator_id IS NOT NULL AND
            authenticator_kind = 'webauthn_v1' AND
            step_up_purpose_sha256 IS NULL AND action_policy_version IS NULL
        ) OR (
            purpose = 'webauthn_authenticate_v1' AND
            proof_kind = 'webauthn_v1' AND
            authenticator_id IS NOT NULL AND
            authenticator_kind = 'webauthn_v1' AND
            step_up_purpose_sha256 IS NULL AND action_policy_version IS NULL
        ) OR (
            purpose = 'totp_enroll_v1' AND
            proof_kind = 'totp_v1' AND session_id IS NOT NULL AND
            authenticator_id IS NOT NULL AND authenticator_kind = 'totp_v1' AND
            step_up_purpose_sha256 IS NULL AND action_policy_version IS NULL
        ) OR (
            purpose = 'reauth_v1' AND session_id IS NOT NULL AND
            step_up_purpose_sha256 IS NOT NULL AND
            pg_catalog.octet_length(step_up_purpose_sha256) = 32 AND
            action_policy_version IS NOT NULL AND
            action_policy_version > 0 AND (
                (
                    proof_kind = 'password_v1' AND
                    authenticator_id IS NULL AND authenticator_kind IS NULL
                ) OR (
                    proof_kind IN ('webauthn_v1', 'totp_v1') AND
                    authenticator_id IS NOT NULL AND
                    authenticator_kind = proof_kind
                )
            )
        )
    ),
    CONSTRAINT iam_auth_challenges_hash_profile_check
        CHECK (hash_profile = 'hmac_sha256_v1'),
    CONSTRAINT iam_auth_challenges_secret_check CHECK (
        pg_catalog.char_length(hmac_key_id) BETWEEN 1 AND 128 AND
        hmac_key_id ~ '^[a-z0-9][a-z0-9._-]*$' AND
        pg_catalog.octet_length(challenge_hmac) = 32 AND
        pg_catalog.octet_length(context_sha256) = 32
    ),
    CONSTRAINT iam_auth_challenges_origin_context_check CHECK (
        (
            proof_kind = 'webauthn_v1' AND
            origin_policy_version IS NOT NULL AND origin_policy_version > 0 AND
            rp_id_sha256 IS NOT NULL AND origin_sha256 IS NOT NULL AND
            pg_catalog.octet_length(rp_id_sha256) = 32 AND
            pg_catalog.octet_length(origin_sha256) = 32 AND
            user_verification_required IS NOT NULL
        ) OR (
            proof_kind IN ('password_v1', 'totp_v1') AND
            origin_policy_version IS NULL AND rp_id_sha256 IS NULL AND
            origin_sha256 IS NULL AND user_verification_required IS NULL
        )
    ),
    CONSTRAINT iam_auth_challenges_attempt_check
        CHECK (attempt_count BETWEEN 0 AND 5),
    CONSTRAINT iam_auth_challenges_state_check CHECK (
        state IN ('pending', 'consumed', 'failed', 'expired', 'cancelled')
    ),
    CONSTRAINT iam_auth_challenges_revision_check CHECK (revision > 0),
    CONSTRAINT iam_auth_challenges_lifecycle_check CHECK (
        (state = 'pending' AND attempt_count < 5 AND terminal_at IS NULL) OR
        (state = 'consumed' AND terminal_at IS NOT NULL AND
            terminal_at <= expires_at) OR
        (state = 'failed' AND attempt_count = 5 AND
            terminal_at IS NOT NULL AND terminal_at <= expires_at) OR
        (state = 'expired' AND terminal_at IS NOT NULL AND
            terminal_at >= expires_at) OR
        (state = 'cancelled' AND terminal_at IS NOT NULL)
    ),
    CONSTRAINT iam_auth_challenges_time_check CHECK (
        updated_at >= created_at AND
        issued_at >= created_at AND issued_at <= updated_at AND
        expires_at > issued_at AND
        expires_at <= issued_at + INTERVAL '5 minutes' AND
        (terminal_at IS NULL OR
            (terminal_at >= issued_at AND terminal_at <= updated_at)) AND
        purge_after >= GREATEST(
            expires_at, COALESCE(terminal_at, expires_at)
        ) + INTERVAL '24 hours'
    )
);

CREATE INDEX iam_auth_challenges_pending_idx
    ON iam_auth_challenges (user_id, state, expires_at)
    WHERE state = 'pending';

CREATE INDEX iam_auth_challenges_retention_idx
    ON iam_auth_challenges (purge_after)
    WHERE state <> 'pending';
