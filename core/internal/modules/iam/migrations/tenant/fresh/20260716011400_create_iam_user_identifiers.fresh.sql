-- +goose Up
CREATE TABLE iam_user_identifiers (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    kind text NOT NULL,
    display_value text NOT NULL,
    normalized_value text COLLATE "C" NOT NULL,
    normalized_sha256 bytea NOT NULL,
    state text NOT NULL,
    revision bigint NOT NULL,
    verified_at timestamptz,
    activated_at timestamptz,
    tombstoned_at timestamptz,
    cooldown_until timestamptz,
    retired_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT iam_user_identifiers_pkey PRIMARY KEY (id),
    CONSTRAINT iam_user_identifiers_identity_key
        UNIQUE (id, user_id, kind),
    CONSTRAINT iam_user_identifiers_acceptance_key
        UNIQUE (id, user_id, kind, normalized_value),
    CONSTRAINT iam_user_identifiers_user_fk FOREIGN KEY (user_id)
        REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_user_identifiers_kind_check
        CHECK (kind IN ('email_v1', 'portal_v1')),
    CONSTRAINT iam_user_identifiers_display_check CHECK (
        pg_catalog.char_length(display_value) BETWEEN 1 AND 320
    ),
    CONSTRAINT iam_user_identifiers_normalized_check CHECK (
        (kind = 'email_v1' AND pg_catalog.octet_length(normalized_value)
            BETWEEN 3 AND 254) OR
        (kind = 'portal_v1' AND pg_catalog.char_length(normalized_value)
            BETWEEN 1 AND 100)
    ),
    CONSTRAINT iam_user_identifiers_normalized_sha256_check
        CHECK (pg_catalog.octet_length(normalized_sha256) = 32),
    CONSTRAINT iam_user_identifiers_state_check
        CHECK (state IN ('pending', 'verified', 'active', 'tombstoned', 'retired')),
    CONSTRAINT iam_user_identifiers_revision_check CHECK (revision > 0),
    CONSTRAINT iam_user_identifiers_timestamp_order_check
        CHECK (updated_at >= created_at),
    CONSTRAINT iam_user_identifiers_lifecycle_check CHECK (
        (state = 'pending' AND verified_at IS NULL AND activated_at IS NULL AND
            tombstoned_at IS NULL AND cooldown_until IS NULL AND retired_at IS NULL) OR
        (state = 'verified' AND verified_at IS NOT NULL AND
            verified_at >= created_at AND activated_at IS NULL AND
            tombstoned_at IS NULL AND cooldown_until IS NULL AND retired_at IS NULL) OR
        (state = 'active' AND verified_at IS NOT NULL AND activated_at IS NOT NULL AND
            verified_at >= created_at AND activated_at >= verified_at AND
            tombstoned_at IS NULL AND cooldown_until IS NULL AND retired_at IS NULL) OR
        (state = 'tombstoned' AND verified_at IS NOT NULL AND activated_at IS NOT NULL AND
            tombstoned_at IS NOT NULL AND cooldown_until IS NOT NULL AND
            verified_at >= created_at AND activated_at >= verified_at AND
            tombstoned_at >= activated_at AND
            cooldown_until = tombstoned_at + INTERVAL '30 days' AND retired_at IS NULL) OR
        (state = 'retired' AND verified_at IS NOT NULL AND activated_at IS NOT NULL AND
            tombstoned_at IS NOT NULL AND cooldown_until IS NOT NULL AND
            retired_at IS NOT NULL AND verified_at >= created_at AND
            activated_at >= verified_at AND tombstoned_at >= activated_at AND
            cooldown_until = tombstoned_at + INTERVAL '30 days' AND
            retired_at >= cooldown_until)
    )
);

CREATE UNIQUE INDEX iam_user_identifiers_reserved_value_uidx
    ON iam_user_identifiers (kind, normalized_value)
    WHERE state <> 'retired';

CREATE UNIQUE INDEX iam_user_identifiers_one_active_uidx
    ON iam_user_identifiers (user_id, kind)
    WHERE state = 'active';
