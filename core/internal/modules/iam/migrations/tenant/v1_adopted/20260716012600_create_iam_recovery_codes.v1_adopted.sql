-- +goose Up
CREATE TABLE iam_recovery_codes (
    id uuid NOT NULL,
    authenticator_id uuid NOT NULL,
    user_id uuid NOT NULL,
    authenticator_kind text NOT NULL,
    position smallint NOT NULL,
    hash_profile text NOT NULL,
    hmac_key_id text COLLATE "C" NOT NULL,
    code_hmac bytea NOT NULL,
    state text NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    consumed_at timestamptz,
    revoked_at timestamptz,
    purge_after timestamptz,
    CONSTRAINT iam_recovery_codes_pkey PRIMARY KEY (id),
    CONSTRAINT iam_recovery_codes_position_key
        UNIQUE (authenticator_id, position),
    CONSTRAINT iam_recovery_codes_digest_key
        UNIQUE (hmac_key_id, code_hmac),
    CONSTRAINT iam_recovery_codes_authenticator_fk
        FOREIGN KEY (authenticator_id, user_id, authenticator_kind)
        REFERENCES iam_authenticators (id, user_id, kind)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_recovery_codes_kind_check
        CHECK (authenticator_kind = 'recovery_set_v1'),
    CONSTRAINT iam_recovery_codes_position_check
        CHECK (position BETWEEN 1 AND 32),
    CONSTRAINT iam_recovery_codes_hash_profile_check
        CHECK (hash_profile = 'hmac_sha256_v1'),
    CONSTRAINT iam_recovery_codes_hmac_key_id_check CHECK (
        pg_catalog.char_length(hmac_key_id) BETWEEN 1 AND 128 AND
        hmac_key_id ~ '^[a-z0-9][a-z0-9._-]*$'
    ),
    CONSTRAINT iam_recovery_codes_hmac_check
        CHECK (pg_catalog.octet_length(code_hmac) = 32),
    CONSTRAINT iam_recovery_codes_state_check
        CHECK (state IN ('active', 'consumed', 'revoked')),
    CONSTRAINT iam_recovery_codes_revision_check CHECK (revision > 0),
    CONSTRAINT iam_recovery_codes_timestamp_check CHECK (
        updated_at >= created_at AND
        (consumed_at IS NULL OR
            (consumed_at >= created_at AND consumed_at <= updated_at)) AND
        (revoked_at IS NULL OR
            (revoked_at >= created_at AND revoked_at <= updated_at))
    ),
    CONSTRAINT iam_recovery_codes_lifecycle_check CHECK (
        (state = 'active' AND consumed_at IS NULL AND
            revoked_at IS NULL AND purge_after IS NULL) OR
        (state = 'consumed' AND consumed_at IS NOT NULL AND
            revoked_at IS NULL AND
            purge_after IS NOT NULL AND
            purge_after = consumed_at + INTERVAL '30 days') OR
        (state = 'revoked' AND consumed_at IS NULL AND
            revoked_at IS NOT NULL AND
            purge_after IS NOT NULL AND
            purge_after = revoked_at + INTERVAL '30 days')
    )
);
