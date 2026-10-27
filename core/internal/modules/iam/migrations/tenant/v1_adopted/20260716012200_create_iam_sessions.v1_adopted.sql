-- +goose Up
CREATE TABLE iam_sessions (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    access_mode_id uuid NOT NULL,
    access_mode text NOT NULL,
    portal_binding_id uuid,
    state text NOT NULL,
    device_kind text COLLATE "C" NOT NULL,
    browser_family text COLLATE "C" NOT NULL,
    os_family text COLLATE "C" NOT NULL,
    ip_hmac_key_id text COLLATE "C" NOT NULL,
    ip_hmac bytea NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    last_seen_at timestamptz NOT NULL,
    absolute_expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    expired_at timestamptz,
    retain_until timestamptz NOT NULL,
    CONSTRAINT iam_sessions_pkey PRIMARY KEY (id),
    CONSTRAINT iam_sessions_user_key UNIQUE (id, user_id),
    CONSTRAINT iam_sessions_user_fk
        FOREIGN KEY (user_id) REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_sessions_access_mode_fk
        FOREIGN KEY (access_mode_id, user_id, access_mode)
        REFERENCES iam_user_access_modes (id, user_id, mode)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_sessions_portal_binding_fk
        FOREIGN KEY (portal_binding_id, user_id, access_mode_id)
        REFERENCES iam_portal_subject_bindings (id, user_id, access_mode_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_sessions_surface_check CHECK (
        (access_mode = 'operator_console' AND portal_binding_id IS NULL) OR
        (access_mode = 'portal' AND portal_binding_id IS NOT NULL)
    ),
    CONSTRAINT iam_sessions_state_check
        CHECK (state IN ('active', 'revoked', 'expired')),
    CONSTRAINT iam_sessions_device_kind_check CHECK (
        device_kind IN ('desktop', 'mobile', 'tablet', 'unknown')
    ),
    CONSTRAINT iam_sessions_device_key_check CHECK (
        pg_catalog.char_length(browser_family) BETWEEN 1 AND 64 AND
        browser_family ~ '^[a-z0-9][a-z0-9._-]*$' AND
        pg_catalog.char_length(os_family) BETWEEN 1 AND 64 AND
        os_family ~ '^[a-z0-9][a-z0-9._-]*$'
    ),
    CONSTRAINT iam_sessions_ip_hmac_check CHECK (
        pg_catalog.char_length(ip_hmac_key_id) BETWEEN 1 AND 128 AND
        ip_hmac_key_id ~ '^[a-z0-9][a-z0-9._-]*$' AND
        pg_catalog.octet_length(ip_hmac) = 32
    ),
    CONSTRAINT iam_sessions_revision_check CHECK (revision > 0),
    CONSTRAINT iam_sessions_lifecycle_check CHECK (
        (state = 'active' AND revoked_at IS NULL AND expired_at IS NULL) OR
        (state = 'revoked' AND revoked_at IS NOT NULL AND expired_at IS NULL) OR
        (state = 'expired' AND revoked_at IS NULL AND
            expired_at = absolute_expires_at)
    ),
    CONSTRAINT iam_sessions_time_check CHECK (
        updated_at >= created_at AND
        last_seen_at >= created_at AND last_seen_at <= updated_at AND
        absolute_expires_at > created_at AND
        absolute_expires_at <= created_at + INTERVAL '90 days' AND
        (revoked_at IS NULL OR
            (revoked_at >= created_at AND revoked_at <= updated_at)) AND
        (expired_at IS NULL OR expired_at <= updated_at) AND
        retain_until >= GREATEST(
            absolute_expires_at,
            COALESCE(revoked_at, absolute_expires_at),
            COALESCE(expired_at, absolute_expires_at)
        ) + INTERVAL '30 days'
    )
);

CREATE INDEX iam_sessions_user_state_idx
    ON iam_sessions (user_id, state, absolute_expires_at);

CREATE INDEX iam_sessions_retention_idx
    ON iam_sessions (retain_until)
    WHERE state <> 'active';

