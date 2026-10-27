-- +goose Up
CREATE TABLE iam_user_role_assignments (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    role_id uuid NOT NULL,
    role_scope_kind text NOT NULL,
    role_scope_id uuid NOT NULL,
    organization_id uuid,
    membership_id uuid,
    state text NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    revoked_at timestamptz,
    CONSTRAINT iam_user_role_assignments_pkey PRIMARY KEY (id),
    CONSTRAINT iam_user_role_assignments_user_role_key UNIQUE (user_id, role_id),
    CONSTRAINT iam_user_role_assignments_user_fk FOREIGN KEY (user_id)
        REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_user_role_assignments_role_fk
        FOREIGN KEY (role_id, role_scope_kind, role_scope_id)
        REFERENCES iam_roles (id, scope_kind, scope_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_user_role_assignments_organization_fk FOREIGN KEY (organization_id)
        REFERENCES iam_organizations (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_user_role_assignments_membership_fk
        FOREIGN KEY (membership_id, organization_id, user_id)
        REFERENCES iam_organization_memberships (id, organization_id, user_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_user_role_assignments_scope_check CHECK (
        (role_scope_kind = 'tenant' AND organization_id IS NULL AND membership_id IS NULL) OR
        (role_scope_kind = 'organization' AND organization_id IS NOT NULL AND
            organization_id = role_scope_id AND membership_id IS NOT NULL)
    ),
    CONSTRAINT iam_user_role_assignments_state_check
        CHECK (state IN ('active', 'suspended', 'revoked')),
    CONSTRAINT iam_user_role_assignments_revision_check CHECK (revision > 0),
    CONSTRAINT iam_user_role_assignments_timestamp_order_check
        CHECK (updated_at >= created_at),
    CONSTRAINT iam_user_role_assignments_revocation_check CHECK (
        (state <> 'revoked' AND revoked_at IS NULL) OR
        (state = 'revoked' AND revoked_at IS NOT NULL AND revoked_at >= updated_at)
    )
);

CREATE INDEX iam_user_role_assignments_role_idx
    ON iam_user_role_assignments (role_id, state);
