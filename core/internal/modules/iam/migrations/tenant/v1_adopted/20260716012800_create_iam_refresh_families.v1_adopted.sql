-- +goose Up
CREATE TABLE iam_refresh_families (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    session_id uuid NOT NULL,
    state text NOT NULL,
    rotation_sequence bigint NOT NULL,
    revoke_reason text COLLATE "C",
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    absolute_expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    compromised_at timestamptz,
    expired_at timestamptz,
    retain_until timestamptz NOT NULL,
    CONSTRAINT iam_refresh_families_pkey PRIMARY KEY (id),
    CONSTRAINT iam_refresh_families_session_key UNIQUE (session_id),
    CONSTRAINT iam_refresh_families_expiry_key
        UNIQUE (id, absolute_expires_at),
    CONSTRAINT iam_refresh_families_user_fk
        FOREIGN KEY (user_id) REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_refresh_families_session_fk
        FOREIGN KEY (session_id, user_id) REFERENCES iam_sessions (id, user_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_refresh_families_state_check CHECK (
        state IN ('active', 'revoked', 'compromised', 'expired')
    ),
    CONSTRAINT iam_refresh_families_counter_check CHECK (
        rotation_sequence >= 0 AND revision > 0
    ),
    CONSTRAINT iam_refresh_families_reason_check CHECK (
        revoke_reason IS NULL OR (
            pg_catalog.char_length(revoke_reason) BETWEEN 1 AND 64 AND
            revoke_reason ~ '^[a-z0-9][a-z0-9._-]*$'
        )
    ),
    CONSTRAINT iam_refresh_families_lifecycle_check CHECK (
        (
            state = 'active' AND revoke_reason IS NULL AND
            revoked_at IS NULL AND compromised_at IS NULL AND expired_at IS NULL
        ) OR (
            state = 'revoked' AND revoke_reason IS NOT NULL AND
            revoked_at IS NOT NULL AND compromised_at IS NULL AND expired_at IS NULL
        ) OR (
            state = 'compromised' AND revoke_reason IS NOT NULL AND
            revoked_at IS NULL AND compromised_at IS NOT NULL AND expired_at IS NULL
        ) OR (
            state = 'expired' AND revoke_reason IS NULL AND
            revoked_at IS NULL AND compromised_at IS NULL AND
            expired_at = absolute_expires_at
        )
    ),
    CONSTRAINT iam_refresh_families_time_check CHECK (
        updated_at >= created_at AND
        absolute_expires_at = created_at + INTERVAL '90 days' AND
        (revoked_at IS NULL OR
            (revoked_at >= created_at AND revoked_at <= updated_at)) AND
        (compromised_at IS NULL OR
            (compromised_at >= created_at AND compromised_at <= updated_at)) AND
        (expired_at IS NULL OR expired_at <= updated_at) AND
        retain_until >= absolute_expires_at + INTERVAL '24 hours'
    )
);

CREATE INDEX iam_refresh_families_user_state_idx
    ON iam_refresh_families (user_id, state, absolute_expires_at);

CREATE INDEX iam_refresh_families_retention_idx
    ON iam_refresh_families (retain_until)
    WHERE state <> 'active';

