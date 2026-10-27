-- +goose Up
CREATE TABLE iam_authenticators (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    kind text NOT NULL,
    display_label text NOT NULL,
    state text NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    activated_at timestamptz,
    last_used_at timestamptz,
    suspended_at timestamptz,
    compromised_at timestamptz,
    renewal_required_at timestamptz,
    retired_at timestamptz,
    CONSTRAINT iam_authenticators_pkey PRIMARY KEY (id),
    CONSTRAINT iam_authenticators_subtype_key UNIQUE (id, user_id, kind),
    CONSTRAINT iam_authenticators_user_fk FOREIGN KEY (user_id)
        REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_authenticators_kind_check
        CHECK (kind IN ('webauthn_v1', 'totp_v1', 'recovery_set_v1')),
    CONSTRAINT iam_authenticators_label_check CHECK (
        pg_catalog.char_length(display_label) BETWEEN 1 AND 100 AND
        display_label = pg_catalog.btrim(display_label)
    ),
    CONSTRAINT iam_authenticators_state_check CHECK (
        state IN (
            'pending', 'active', 'suspended', 'compromised',
            'renewal_required', 'retired'
        )
    ),
    CONSTRAINT iam_authenticators_revision_check CHECK (revision > 0),
    CONSTRAINT iam_authenticators_timestamp_check CHECK (
        updated_at >= created_at AND
        (activated_at IS NULL OR
            (activated_at >= created_at AND activated_at <= updated_at)) AND
        (last_used_at IS NULL OR
            (activated_at IS NOT NULL AND last_used_at >= activated_at AND
                last_used_at <= updated_at)) AND
        (suspended_at IS NULL OR
            (suspended_at >= COALESCE(activated_at, created_at) AND
                suspended_at <= updated_at)) AND
        (compromised_at IS NULL OR
            (compromised_at >= COALESCE(activated_at, created_at) AND
                compromised_at <= updated_at)) AND
        (renewal_required_at IS NULL OR
            (activated_at IS NOT NULL AND
                renewal_required_at >= activated_at AND
                renewal_required_at <= updated_at)) AND
        (retired_at IS NULL OR
            (retired_at >= created_at AND retired_at <= updated_at))
    ),
    CONSTRAINT iam_authenticators_lifecycle_check CHECK (
        (state = 'pending' AND activated_at IS NULL AND
            suspended_at IS NULL AND compromised_at IS NULL AND
            renewal_required_at IS NULL AND retired_at IS NULL) OR
        (state = 'active' AND activated_at IS NOT NULL AND
            suspended_at IS NULL AND compromised_at IS NULL AND
            renewal_required_at IS NULL AND retired_at IS NULL) OR
        (state = 'suspended' AND suspended_at IS NOT NULL AND
            compromised_at IS NULL AND renewal_required_at IS NULL AND
            retired_at IS NULL) OR
        (state = 'compromised' AND compromised_at IS NOT NULL AND
            suspended_at IS NULL AND renewal_required_at IS NULL AND
            retired_at IS NULL) OR
        (state = 'renewal_required' AND activated_at IS NOT NULL AND
            renewal_required_at IS NOT NULL AND suspended_at IS NULL AND
            compromised_at IS NULL AND retired_at IS NULL) OR
        (state = 'retired' AND retired_at IS NOT NULL AND
            (suspended_at IS NULL OR retired_at >= suspended_at) AND
            (compromised_at IS NULL OR retired_at >= compromised_at) AND
            (renewal_required_at IS NULL OR
                retired_at >= renewal_required_at))
    )
);

CREATE UNIQUE INDEX iam_authenticators_one_live_totp_uidx
    ON iam_authenticators (user_id)
    WHERE kind = 'totp_v1' AND state <> 'retired';

CREATE UNIQUE INDEX iam_authenticators_one_live_recovery_set_uidx
    ON iam_authenticators (user_id)
    WHERE kind = 'recovery_set_v1' AND state <> 'retired';

CREATE INDEX iam_authenticators_user_state_idx
    ON iam_authenticators (user_id, state, kind);

