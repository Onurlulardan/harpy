-- +goose Up
ALTER TABLE iam_users
    ADD COLUMN identity_kind text;

UPDATE iam_users
SET identity_kind = 'human';

ALTER TABLE iam_users
    ALTER COLUMN identity_kind SET NOT NULL,
    ADD CONSTRAINT iam_users_identity_kind_check
        CHECK (identity_kind = 'human');

CREATE TABLE iam_user_access_modes (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    mode text NOT NULL,
    state text NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    activated_at timestamptz,
    suspended_at timestamptz,
    retired_at timestamptz,
    CONSTRAINT iam_user_access_modes_pkey PRIMARY KEY (id),
    CONSTRAINT iam_user_access_modes_user_mode_key UNIQUE (user_id, mode),
    CONSTRAINT iam_user_access_modes_binding_key UNIQUE (id, user_id, mode),
    CONSTRAINT iam_user_access_modes_user_fk
        FOREIGN KEY (user_id) REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_user_access_modes_mode_check
        CHECK (mode IN ('operator_console', 'portal')),
    CONSTRAINT iam_user_access_modes_state_check
        CHECK (state IN ('pending', 'active', 'suspended', 'retired')),
    CONSTRAINT iam_user_access_modes_revision_check CHECK (revision > 0),
    CONSTRAINT iam_user_access_modes_timestamp_check CHECK (
        updated_at >= created_at AND
        (activated_at IS NULL OR
            (activated_at >= created_at AND activated_at <= updated_at)) AND
        (suspended_at IS NULL OR
            (suspended_at >= COALESCE(activated_at, created_at) AND
                suspended_at <= updated_at)) AND
        (retired_at IS NULL OR
            (retired_at >= created_at AND retired_at <= updated_at))
    ),
    CONSTRAINT iam_user_access_modes_lifecycle_check CHECK (
        (state = 'pending' AND activated_at IS NULL AND
            suspended_at IS NULL AND retired_at IS NULL) OR
        (state = 'active' AND activated_at IS NOT NULL AND
            suspended_at IS NULL AND retired_at IS NULL) OR
        (state = 'suspended' AND suspended_at IS NOT NULL AND
            retired_at IS NULL) OR
        (state = 'retired' AND retired_at IS NOT NULL AND
            (suspended_at IS NULL OR retired_at >= suspended_at) AND
            (activated_at IS NULL OR retired_at >= activated_at))
    )
);

INSERT INTO iam_user_access_modes (
    id, user_id, mode, state, revision, created_at, updated_at,
    activated_at, suspended_at, retired_at
)
SELECT
    pg_catalog.gen_random_uuid(),
    id,
    CASE account_type
        WHEN 'operator' THEN 'operator_console'
        WHEN 'portal' THEN 'portal'
    END,
    'active',
    1,
    created_at,
    updated_at,
    created_at,
    NULL,
    NULL
FROM iam_users;

CREATE INDEX iam_user_access_modes_admission_idx
    ON iam_user_access_modes (mode, state, user_id);

ALTER TABLE iam_users
    DROP CONSTRAINT iam_users_account_type_check,
    DROP COLUMN account_type;

CREATE INDEX iam_users_status_idx
    ON iam_users (status, identity_kind);
