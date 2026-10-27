-- +goose Up
CREATE TABLE iam_webauthn_credentials (
    authenticator_id uuid NOT NULL,
    user_id uuid NOT NULL,
    authenticator_kind text NOT NULL,
    credential_id bytea NOT NULL,
    public_key_cose bytea NOT NULL,
    rp_id text COLLATE "C" NOT NULL,
    rp_id_sha256 bytea NOT NULL,
    origin_policy_version bigint NOT NULL,
    aaguid uuid NOT NULL,
    sign_count bigint NOT NULL,
    backup_eligible boolean NOT NULL,
    backup_state boolean NOT NULL,
    discoverable boolean NOT NULL,
    uv_required boolean NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT iam_webauthn_credentials_pkey PRIMARY KEY (authenticator_id),
    CONSTRAINT iam_webauthn_credentials_credential_id_key UNIQUE (credential_id),
    CONSTRAINT iam_webauthn_credentials_authenticator_fk
        FOREIGN KEY (authenticator_id, user_id, authenticator_kind)
        REFERENCES iam_authenticators (id, user_id, kind)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_webauthn_credentials_kind_check
        CHECK (authenticator_kind = 'webauthn_v1'),
    CONSTRAINT iam_webauthn_credentials_credential_id_check CHECK (
        pg_catalog.octet_length(credential_id) BETWEEN 1 AND 1024
    ),
    CONSTRAINT iam_webauthn_credentials_public_key_check CHECK (
        pg_catalog.octet_length(public_key_cose) BETWEEN 1 AND 4096
    ),
    CONSTRAINT iam_webauthn_credentials_rp_id_check CHECK (
        pg_catalog.char_length(rp_id) BETWEEN 1 AND 253 AND
        rp_id ~ '^[a-z0-9.-]+$' AND
        pg_catalog.left(rp_id, 1) <> '.' AND
        pg_catalog.right(rp_id, 1) <> '.' AND
        pg_catalog.strpos(rp_id, '..') = 0
    ),
    CONSTRAINT iam_webauthn_credentials_binding_check CHECK (
        pg_catalog.octet_length(rp_id_sha256) = 32 AND
        origin_policy_version > 0
    ),
    CONSTRAINT iam_webauthn_credentials_counter_check CHECK (sign_count >= 0),
    CONSTRAINT iam_webauthn_credentials_backup_check
        CHECK (NOT backup_state OR backup_eligible),
    CONSTRAINT iam_webauthn_credentials_revision_check CHECK (revision > 0),
    CONSTRAINT iam_webauthn_credentials_timestamp_check
        CHECK (updated_at >= created_at)
);
