-- +goose Up
CREATE TABLE notifications_provider_secrets (
    id uuid PRIMARY KEY,
    ciphertext bytea NOT NULL CHECK (pg_catalog.octet_length(ciphertext) BETWEEN 17 AND 32768),
    nonce bytea NOT NULL CHECK (pg_catalog.octet_length(nonce) = 12),
    wrapped_dek bytea NOT NULL CHECK (pg_catalog.octet_length(wrapped_dek) BETWEEN 28 AND 4096),
    kek_revision bigint NOT NULL CHECK (kek_revision > 0),
    plaintext_digest bytea NOT NULL CHECK (pg_catalog.octet_length(plaintext_digest) = 32),
    created_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(created_at)),
    rewrapped_at timestamptz
);

CREATE TABLE notifications_provider_configs (
    id uuid PRIMARY KEY,
    provider_kind text COLLATE "C" NOT NULL CHECK (provider_kind IN ('smtp','telegram')),
    name text NOT NULL CHECK (pg_catalog.octet_length(name) BETWEEN 1 AND 512),
    enabled boolean NOT NULL,
    current_version_id uuid,
    row_version bigint NOT NULL CHECK (row_version > 0),
    created_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(created_at)),
    updated_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(updated_at))
);

CREATE UNIQUE INDEX notifications_provider_configs_one_enabled_kind
    ON notifications_provider_configs (provider_kind) WHERE enabled;

CREATE TABLE notifications_provider_config_versions (
    id uuid PRIMARY KEY,
    config_id uuid NOT NULL,
    version_number bigint NOT NULL CHECK (version_number > 0),
    secret_id uuid NOT NULL,
    metadata jsonb NOT NULL CHECK (pg_catalog.jsonb_typeof(metadata) = 'object'),
    metadata_digest bytea NOT NULL CHECK (pg_catalog.octet_length(metadata_digest) = 32),
    created_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(created_at)),
    CONSTRAINT notifications_provider_config_versions_release_key
        UNIQUE (config_id, version_number),
    CONSTRAINT notifications_provider_config_versions_config_fk FOREIGN KEY (config_id)
        REFERENCES notifications_provider_configs(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT notifications_provider_config_versions_secret_fk FOREIGN KEY (secret_id)
        REFERENCES notifications_provider_secrets(id) ON UPDATE NO ACTION ON DELETE RESTRICT
);

ALTER TABLE notifications_provider_configs
    ADD CONSTRAINT notifications_provider_configs_current_version_fk
    FOREIGN KEY (current_version_id) REFERENCES notifications_provider_config_versions(id)
    ON UPDATE NO ACTION ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE notifications_deliveries
    ADD CONSTRAINT notifications_deliveries_provider_version_fk
    FOREIGN KEY (provider_config_version_id) REFERENCES notifications_provider_config_versions(id)
    ON UPDATE NO ACTION ON DELETE RESTRICT;

CREATE INDEX notifications_provider_config_versions_lookup_idx
    ON notifications_provider_config_versions (config_id, version_number DESC);
