-- +goose Up
CREATE TABLE iam_totp_authenticators (
    authenticator_id uuid NOT NULL,
    user_id uuid NOT NULL,
    authenticator_kind text NOT NULL,
    profile text NOT NULL,
    cipher_suite text NOT NULL,
    kms_key_id text COLLATE "C" NOT NULL,
    nonce bytea NOT NULL,
    secret_ciphertext bytea NOT NULL,
    authentication_tag bytea NOT NULL,
    aad_sha256 bytea NOT NULL,
    last_accepted_time_step bigint,
    verified_at timestamptz,
    enrollment_expires_at timestamptz NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT iam_totp_authenticators_pkey PRIMARY KEY (authenticator_id),
    CONSTRAINT iam_totp_authenticators_nonce_key UNIQUE (kms_key_id, nonce),
    CONSTRAINT iam_totp_authenticators_authenticator_fk
        FOREIGN KEY (authenticator_id, user_id, authenticator_kind)
        REFERENCES iam_authenticators (id, user_id, kind)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_totp_authenticators_kind_check
        CHECK (authenticator_kind = 'totp_v1'),
    CONSTRAINT iam_totp_authenticators_profile_check
        CHECK (profile = 'totp_sha256_6_30_v1'),
    CONSTRAINT iam_totp_authenticators_cipher_suite_check
        CHECK (cipher_suite = 'aes_256_gcm_v1'),
    CONSTRAINT iam_totp_authenticators_kms_key_id_check CHECK (
        pg_catalog.char_length(kms_key_id) BETWEEN 1 AND 128 AND
        kms_key_id ~ '^[a-z0-9][a-z0-9._-]*$'
    ),
    CONSTRAINT iam_totp_authenticators_secret_envelope_check CHECK (
        pg_catalog.octet_length(nonce) = 12 AND
        pg_catalog.octet_length(secret_ciphertext) = 32 AND
        pg_catalog.octet_length(authentication_tag) = 16 AND
        pg_catalog.octet_length(aad_sha256) = 32
    ),
    CONSTRAINT iam_totp_authenticators_replay_check CHECK (
        last_accepted_time_step IS NULL OR last_accepted_time_step >= 0
    ),
    CONSTRAINT iam_totp_authenticators_enrollment_check CHECK (
        enrollment_expires_at = created_at + INTERVAL '10 minutes' AND
        (verified_at IS NULL OR
            (verified_at >= created_at AND
                verified_at <= enrollment_expires_at AND
                verified_at <= updated_at))
    ),
    CONSTRAINT iam_totp_authenticators_revision_check CHECK (revision > 0),
    CONSTRAINT iam_totp_authenticators_timestamp_check
        CHECK (updated_at >= created_at)
);
