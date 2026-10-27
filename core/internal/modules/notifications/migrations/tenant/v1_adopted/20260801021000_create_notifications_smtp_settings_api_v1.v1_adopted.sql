-- +goose Up
CREATE TABLE notifications_smtp_settings_receipts (
    idempotency_key text COLLATE "C" PRIMARY KEY
        CHECK (idempotency_key ~ '^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$'),
    request_digest bytea NOT NULL
        CHECK (pg_catalog.octet_length(request_digest) = 32),
    expected_settings_revision bigint NOT NULL
        CHECK (expected_settings_revision > 0),
    result_settings_revision bigint NOT NULL UNIQUE,
    provider_config_id uuid NOT NULL,
    provider_revision bigint NOT NULL CHECK (provider_revision > 0),
    created_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(created_at)),
    CONSTRAINT notifications_smtp_settings_result_revision_check
        CHECK (result_settings_revision = expected_settings_revision + 1),
    CONSTRAINT notifications_smtp_settings_provider_fk
        FOREIGN KEY (provider_config_id) REFERENCES notifications_provider_configs(id)
        ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE TABLE notifications_smtp_test_intents (
    operation_id uuid PRIMARY KEY,
    idempotency_key text COLLATE "C" NOT NULL UNIQUE
        CHECK (idempotency_key ~ '^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$'),
    request_digest bytea NOT NULL
        CHECK (pg_catalog.octet_length(request_digest) = 32),
    expected_settings_revision bigint NOT NULL
        CHECK (expected_settings_revision > 0),
    provider_config_version_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    target_ciphertext bytea NOT NULL
        CHECK (pg_catalog.octet_length(target_ciphertext) BETWEEN 17 AND 32768),
    target_nonce bytea NOT NULL
        CHECK (pg_catalog.octet_length(target_nonce) = 12),
    target_wrapped_dek bytea NOT NULL
        CHECK (pg_catalog.octet_length(target_wrapped_dek) BETWEEN 28 AND 4096),
    target_kek_revision bigint NOT NULL CHECK (target_kek_revision > 0),
    target_digest bytea NOT NULL
        CHECK (pg_catalog.octet_length(target_digest) = 32),
    lifecycle_state text COLLATE "C" NOT NULL DEFAULT 'queued'
        CHECK (lifecycle_state = 'queued'),
    requested_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(requested_at)),
    created_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(created_at)),
    CONSTRAINT notifications_smtp_test_provider_version_fk
        FOREIGN KEY (provider_config_version_id)
        REFERENCES notifications_provider_config_versions(id)
        ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE TABLE notifications_smtp_settings_lock (
    singleton boolean PRIMARY KEY CHECK (singleton),
    created_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(created_at))
);

INSERT INTO notifications_smtp_settings_lock (singleton, created_at)
VALUES (true, pg_catalog.transaction_timestamp());

-- +goose StatementBegin
CREATE FUNCTION notifications_smtp_immutable_v1()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $$
BEGIN
    RAISE EXCEPTION USING
        ERRCODE = '55000',
        MESSAGE = 'notification SMTP receipt or intent is immutable';
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER notifications_smtp_settings_receipts_immutable
BEFORE UPDATE OR DELETE ON notifications_smtp_settings_receipts
FOR EACH ROW EXECUTE FUNCTION notifications_smtp_immutable_v1();

CREATE TRIGGER notifications_smtp_test_intents_immutable
BEFORE UPDATE OR DELETE ON notifications_smtp_test_intents
FOR EACH ROW EXECUTE FUNCTION notifications_smtp_immutable_v1();

REVOKE ALL ON FUNCTION notifications_smtp_immutable_v1() FROM PUBLIC;

GRANT SELECT ON TABLE notifications_smtp_settings_lock TO harpy_runtime;
GRANT UPDATE (singleton) ON notifications_smtp_settings_lock TO harpy_runtime;
GRANT SELECT ON TABLE notifications_smtp_settings_receipts TO harpy_runtime;
GRANT INSERT (
    idempotency_key, request_digest, expected_settings_revision,
    result_settings_revision, provider_config_id, provider_revision, created_at
) ON notifications_smtp_settings_receipts TO harpy_runtime;
GRANT SELECT ON TABLE notifications_smtp_test_intents TO harpy_runtime;
GRANT INSERT (
    operation_id, idempotency_key, request_digest, expected_settings_revision,
    provider_config_version_id, actor_id, target_ciphertext, target_nonce,
    target_wrapped_dek, target_kek_revision, target_digest, requested_at, created_at
) ON notifications_smtp_test_intents TO harpy_runtime;
