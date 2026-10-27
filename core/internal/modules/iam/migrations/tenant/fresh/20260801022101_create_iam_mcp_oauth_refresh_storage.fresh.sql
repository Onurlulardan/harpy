-- +goose Up
CREATE TABLE iam_mcp_oauth_refresh_families (
    id uuid PRIMARY KEY,
    grant_id uuid NOT NULL UNIQUE,
    state text COLLATE "C" NOT NULL,
    rotation_sequence bigint NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    absolute_expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    compromised_at timestamptz,
    retain_until timestamptz NOT NULL,
    CONSTRAINT iam_mcp_oauth_refresh_families_grant_fk
        FOREIGN KEY (grant_id) REFERENCES iam_mcp_oauth_grants (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_mcp_oauth_refresh_families_state_ck CHECK (
        state IN ('active','revoked','compromised','expired') AND
        rotation_sequence > 0 AND revision > 0 AND
        ((state = 'active' AND revoked_at IS NULL AND compromised_at IS NULL) OR
         (state = 'revoked' AND revoked_at IS NOT NULL AND compromised_at IS NULL) OR
         (state = 'compromised' AND revoked_at IS NULL AND compromised_at IS NOT NULL) OR
         (state = 'expired' AND revoked_at IS NULL AND compromised_at IS NULL))
    ),
    CONSTRAINT iam_mcp_oauth_refresh_families_time_ck CHECK (
        updated_at >= created_at AND
        absolute_expires_at = created_at + INTERVAL '90 days' AND
        (revoked_at IS NULL OR revoked_at BETWEEN created_at AND updated_at) AND
        (compromised_at IS NULL OR compromised_at BETWEEN created_at AND updated_at) AND
        retain_until >= absolute_expires_at + INTERVAL '24 hours'
    )
);

CREATE INDEX iam_mcp_oauth_refresh_families_retention_idx
    ON iam_mcp_oauth_refresh_families (retain_until);

CREATE TABLE iam_mcp_oauth_refresh_tokens (
    id uuid PRIMARY KEY,
    family_id uuid NOT NULL,
    parent_id uuid,
    sequence bigint NOT NULL,
    selector bytea NOT NULL UNIQUE,
    hash_key_id text COLLATE "C" NOT NULL,
    secret_hmac bytea NOT NULL,
    successor_id uuid,
    state text COLLATE "C" NOT NULL,
    revision bigint NOT NULL,
    issued_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    idle_expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    revoked_at timestamptz,
    retain_until timestamptz NOT NULL,
    CONSTRAINT iam_mcp_oauth_refresh_tokens_family_sequence_key
        UNIQUE (family_id, sequence),
    CONSTRAINT iam_mcp_oauth_refresh_tokens_secret_key
        UNIQUE (hash_key_id, secret_hmac),
    CONSTRAINT iam_mcp_oauth_refresh_tokens_family_fk
        FOREIGN KEY (family_id) REFERENCES iam_mcp_oauth_refresh_families (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_mcp_oauth_refresh_tokens_parent_fk
        FOREIGN KEY (parent_id) REFERENCES iam_mcp_oauth_refresh_tokens (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_mcp_oauth_refresh_tokens_successor_fk
        FOREIGN KEY (successor_id) REFERENCES iam_mcp_oauth_refresh_tokens (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT
        DEFERRABLE INITIALLY DEFERRED,
    CONSTRAINT iam_mcp_oauth_refresh_tokens_identity_ck CHECK (
        sequence > 0 AND
        ((sequence = 1 AND parent_id IS NULL) OR
         (sequence > 1 AND parent_id IS NOT NULL AND parent_id <> id)) AND
        pg_catalog.octet_length(selector) = 16 AND
        pg_catalog.char_length(hash_key_id) BETWEEN 1 AND 128 AND
        pg_catalog.octet_length(secret_hmac) = 32 AND revision > 0
    ),
    CONSTRAINT iam_mcp_oauth_refresh_tokens_state_ck CHECK (
        state IN ('active','consumed','revoked','expired') AND
        ((state = 'active' AND consumed_at IS NULL AND revoked_at IS NULL AND
            successor_id IS NULL) OR
         (state = 'consumed' AND consumed_at IS NOT NULL AND revoked_at IS NULL AND
            successor_id IS NOT NULL) OR
         (state = 'revoked' AND consumed_at IS NULL AND revoked_at IS NOT NULL AND
            successor_id IS NULL) OR
         (state = 'expired' AND consumed_at IS NULL AND revoked_at IS NULL AND
            successor_id IS NULL))
    ),
    CONSTRAINT iam_mcp_oauth_refresh_tokens_time_ck CHECK (
        updated_at >= issued_at AND idle_expires_at > issued_at AND
        idle_expires_at <= issued_at + INTERVAL '30 days' AND
        (consumed_at IS NULL OR consumed_at BETWEEN issued_at AND updated_at) AND
        (revoked_at IS NULL OR revoked_at BETWEEN issued_at AND updated_at) AND
        retain_until >= idle_expires_at + INTERVAL '24 hours'
    )
);

CREATE UNIQUE INDEX iam_mcp_oauth_refresh_tokens_one_active_family_uidx
    ON iam_mcp_oauth_refresh_tokens (family_id) WHERE state = 'active';
CREATE INDEX iam_mcp_oauth_refresh_tokens_retention_idx
    ON iam_mcp_oauth_refresh_tokens (retain_until);

ALTER TABLE iam_mcp_oauth_access_tokens
    ADD COLUMN refresh_token_id uuid NOT NULL,
    ADD CONSTRAINT iam_mcp_oauth_access_tokens_refresh_key
        UNIQUE (refresh_token_id),
    ADD CONSTRAINT iam_mcp_oauth_access_tokens_refresh_fk
        FOREIGN KEY (refresh_token_id)
        REFERENCES iam_mcp_oauth_refresh_tokens (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT;
