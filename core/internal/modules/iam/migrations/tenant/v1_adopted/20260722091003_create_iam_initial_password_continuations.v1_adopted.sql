-- +goose Up
CREATE TABLE iam_initial_password_continuations (
    selector bytea NOT NULL,
    tenant_id uuid NOT NULL,
    user_id uuid NOT NULL,
    installation_operation_id uuid NOT NULL,
    credential_id uuid NOT NULL,
    credential_version bigint NOT NULL,
    credential_revision bigint NOT NULL,
    secret_key_id text COLLATE "C" NOT NULL,
    secret_hmac bytea NOT NULL,
    csrf_key_id text COLLATE "C" NOT NULL,
    csrf_hmac bytea NOT NULL,
    state text NOT NULL,
    policy_attempts smallint NOT NULL,
    revision bigint NOT NULL,
    issued_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    consumed_at timestamptz,
    CONSTRAINT iam_initial_password_continuations_pkey PRIMARY KEY (selector),
    CONSTRAINT iam_initial_password_continuations_user_fk
        FOREIGN KEY (user_id) REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_initial_password_continuations_credential_fk
        FOREIGN KEY (credential_id)
        REFERENCES iam_installation_password_credentials (credential_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_initial_password_continuations_selector_check
        CHECK (pg_catalog.octet_length(selector) OPERATOR(pg_catalog.=) 16),
    CONSTRAINT iam_initial_password_continuations_version_check CHECK (
        credential_version OPERATOR(pg_catalog.>) 0 AND
        credential_revision OPERATOR(pg_catalog.>) 0 AND
        revision OPERATOR(pg_catalog.>) 0
    ),
    CONSTRAINT iam_initial_password_continuations_key_check CHECK (
        secret_key_id OPERATOR(pg_catalog.~) '^[A-Za-z0-9_-]{16,64}$' AND
        csrf_key_id OPERATOR(pg_catalog.~) '^[A-Za-z0-9_-]{16,64}$' AND
        secret_key_id OPERATOR(pg_catalog.<>) csrf_key_id AND
        pg_catalog.octet_length(secret_hmac) OPERATOR(pg_catalog.=) 32 AND
        pg_catalog.octet_length(csrf_hmac) OPERATOR(pg_catalog.=) 32
    ),
    CONSTRAINT iam_initial_password_continuations_state_check CHECK (
        state OPERATOR(pg_catalog.=) ANY (ARRAY[
            'active'::text, 'superseded'::text, 'consumed'::text,
            'exhausted'::text, 'expired'::text
        ])
    ),
    CONSTRAINT iam_initial_password_continuations_attempt_check
        CHECK (policy_attempts BETWEEN 0 AND 5),
    CONSTRAINT iam_initial_password_continuations_lifetime_check CHECK (
        expires_at OPERATOR(pg_catalog.=) (
            issued_at OPERATOR(pg_catalog.+) INTERVAL '10 minutes') AND
        updated_at OPERATOR(pg_catalog.>=) issued_at AND
        (consumed_at IS NULL OR consumed_at OPERATOR(pg_catalog.>=) issued_at)
    ),
    CONSTRAINT iam_initial_password_continuations_state_shape_check CHECK (
        (state OPERATOR(pg_catalog.=) 'active' AND consumed_at IS NULL AND
            policy_attempts OPERATOR(pg_catalog.<) 5) OR
        (state OPERATOR(pg_catalog.=) 'exhausted' AND consumed_at IS NOT NULL AND
            policy_attempts OPERATOR(pg_catalog.=) 5) OR
        (state OPERATOR(pg_catalog.=) ANY (
            ARRAY['superseded'::text, 'consumed'::text, 'expired'::text]
        ) AND consumed_at IS NOT NULL)
    )
);

CREATE UNIQUE INDEX iam_initial_password_continuations_active_user_unique
    ON iam_initial_password_continuations (user_id)
    WHERE state OPERATOR(pg_catalog.=) 'active';

CREATE INDEX iam_initial_password_continuations_expiry_idx
    ON iam_initial_password_continuations (expires_at)
    WHERE state OPERATOR(pg_catalog.=) 'active';
