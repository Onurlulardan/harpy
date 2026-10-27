-- +goose Up
CREATE TABLE iam_organization_memberships (
    id uuid NOT NULL,
    organization_id uuid NOT NULL,
    user_id uuid NOT NULL,
    membership_kind text NOT NULL,
    state text NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    tombstoned_at timestamptz,
    CONSTRAINT iam_organization_memberships_pkey PRIMARY KEY (id),
    CONSTRAINT iam_organization_memberships_relation_key
        UNIQUE (organization_id, user_id),
    CONSTRAINT iam_organization_memberships_assignment_key
        UNIQUE (id, organization_id, user_id),
    CONSTRAINT iam_organization_memberships_organization_fk
        FOREIGN KEY (organization_id) REFERENCES iam_organizations (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_organization_memberships_user_fk FOREIGN KEY (user_id)
        REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_organization_memberships_kind_check
        CHECK (membership_kind IN ('member', 'owner')),
    CONSTRAINT iam_organization_memberships_state_check
        CHECK (state IN ('active', 'suspended', 'tombstoned')),
    CONSTRAINT iam_organization_memberships_revision_check CHECK (revision > 0),
    CONSTRAINT iam_organization_memberships_timestamp_order_check
        CHECK (updated_at >= created_at),
    CONSTRAINT iam_organization_memberships_tombstone_check CHECK (
        (state <> 'tombstoned' AND tombstoned_at IS NULL) OR
        (state = 'tombstoned' AND tombstoned_at IS NOT NULL AND
            tombstoned_at >= updated_at)
    )
);

CREATE INDEX iam_organization_memberships_user_idx
    ON iam_organization_memberships (user_id, state);
