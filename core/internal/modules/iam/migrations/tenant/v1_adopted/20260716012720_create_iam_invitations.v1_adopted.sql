-- +goose Up
CREATE TABLE iam_invitations (
    id uuid NOT NULL,
    identifier_kind text NOT NULL,
    display_value text NOT NULL,
    normalized_value text COLLATE "C" NOT NULL,
    normalized_sha256 bytea NOT NULL,
    role_id uuid NOT NULL,
    role_scope_kind text NOT NULL,
    role_scope_id uuid NOT NULL,
    organization_id uuid,
    invited_by uuid NOT NULL,
    accepted_user_id uuid,
    accepted_identifier_id uuid,
    accepted_identifier_kind text,
    state text NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    accepted_at timestamptz,
    revoked_at timestamptz,
    expired_at timestamptz,
    CONSTRAINT iam_invitations_pkey PRIMARY KEY (id),
    CONSTRAINT iam_invitations_expiry_key UNIQUE (id, expires_at),
    CONSTRAINT iam_invitations_role_fk
        FOREIGN KEY (role_id, role_scope_kind, role_scope_id)
        REFERENCES iam_roles (id, scope_kind, scope_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_invitations_organization_fk
        FOREIGN KEY (organization_id) REFERENCES iam_organizations (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_invitations_invited_by_fk
        FOREIGN KEY (invited_by) REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_invitations_accepted_identifier_fk FOREIGN KEY (
        accepted_identifier_id, accepted_user_id,
        accepted_identifier_kind, normalized_value
    ) REFERENCES iam_user_identifiers (
        id, user_id, kind, normalized_value
    )
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_invitations_identifier_kind_check
        CHECK (identifier_kind = 'email_v1'),
    CONSTRAINT iam_invitations_accepted_identifier_kind_check CHECK (
        accepted_identifier_kind IS NULL OR accepted_identifier_kind = 'email_v1'
    ),
    CONSTRAINT iam_invitations_identifier_check CHECK (
        pg_catalog.char_length(display_value) BETWEEN 1 AND 320 AND
        pg_catalog.char_length(normalized_value) BETWEEN 3 AND 254 AND
        pg_catalog.octet_length(normalized_sha256) = 32
    ),
    CONSTRAINT iam_invitations_role_scope_check CHECK (
        (
            role_scope_kind = 'tenant' AND organization_id IS NULL
        ) OR (
            role_scope_kind = 'organization' AND
            organization_id IS NOT NULL AND
            organization_id = role_scope_id
        )
    ),
    CONSTRAINT iam_invitations_state_check
        CHECK (state IN ('pending', 'accepted', 'revoked', 'expired')),
    CONSTRAINT iam_invitations_revision_check CHECK (revision > 0),
    CONSTRAINT iam_invitations_lifecycle_check CHECK (
        (
            state = 'pending' AND accepted_user_id IS NULL AND
            accepted_identifier_id IS NULL AND accepted_identifier_kind IS NULL AND
            accepted_at IS NULL AND revoked_at IS NULL AND expired_at IS NULL
        ) OR (
            state = 'accepted' AND accepted_user_id IS NOT NULL AND
            accepted_identifier_id IS NOT NULL AND
            accepted_identifier_kind = 'email_v1' AND
            accepted_at IS NOT NULL AND revoked_at IS NULL AND expired_at IS NULL
        ) OR (
            state = 'revoked' AND accepted_user_id IS NULL AND
            accepted_identifier_id IS NULL AND accepted_identifier_kind IS NULL AND
            accepted_at IS NULL AND revoked_at IS NOT NULL AND expired_at IS NULL
        ) OR (
            state = 'expired' AND accepted_user_id IS NULL AND
            accepted_identifier_id IS NULL AND accepted_identifier_kind IS NULL AND
            accepted_at IS NULL AND revoked_at IS NULL AND
            expired_at = expires_at
        )
    ),
    CONSTRAINT iam_invitations_time_check CHECK (
        updated_at >= created_at AND
        expires_at = created_at + INTERVAL '7 days' AND
        (accepted_at IS NULL OR
            (accepted_at >= created_at AND
                accepted_at <= expires_at AND accepted_at <= updated_at)) AND
        (revoked_at IS NULL OR
            (revoked_at >= created_at AND revoked_at <= updated_at)) AND
        (expired_at IS NULL OR expired_at <= updated_at)
    )
);

CREATE UNIQUE INDEX iam_invitations_one_pending_target_uidx
    ON iam_invitations (identifier_kind, normalized_value, role_id)
    WHERE state = 'pending';

CREATE INDEX iam_invitations_expiry_idx
    ON iam_invitations (state, expires_at)
    WHERE state = 'pending';
