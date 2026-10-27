-- +goose Up
CREATE TABLE iam_permission_requests (
    id uuid NOT NULL,
    group_id uuid NOT NULL,
    ordinal smallint NOT NULL,
    target_kind text NOT NULL,
    target_user_id uuid,
    target_role_id uuid,
    target_organization_id uuid,
    permission_id uuid NOT NULL,
    permission_match_kind text NOT NULL,
    permission_is_protected boolean NOT NULL,
    planned_binding_id uuid NOT NULL,
    granted_binding_id uuid,
    state text NOT NULL,
    revision bigint NOT NULL,
    terminal_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT iam_permission_requests_pkey PRIMARY KEY (id),
    CONSTRAINT iam_permission_requests_group_ordinal_key UNIQUE (group_id, ordinal),
    CONSTRAINT iam_permission_requests_planned_binding_key UNIQUE (planned_binding_id),
    CONSTRAINT iam_permission_requests_group_state_fk
        FOREIGN KEY (group_id, state)
        REFERENCES iam_permission_request_groups (id, state)
        ON UPDATE NO ACTION ON DELETE RESTRICT
        DEFERRABLE INITIALLY DEFERRED,
    CONSTRAINT iam_permission_requests_user_fk FOREIGN KEY (target_user_id)
        REFERENCES iam_users (id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_permission_requests_role_fk FOREIGN KEY (target_role_id)
        REFERENCES iam_roles (id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_permission_requests_organization_fk
        FOREIGN KEY (target_organization_id) REFERENCES iam_organizations (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_permission_requests_permission_fk
        FOREIGN KEY (permission_id, permission_match_kind, permission_is_protected)
        REFERENCES iam_permissions (id, match_kind, is_protected)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_permission_requests_ordinal_check
        CHECK (ordinal BETWEEN 1 AND 100),
    CONSTRAINT iam_permission_requests_target_kind_check
        CHECK (target_kind IN ('USER', 'ROLE', 'ORGANIZATION')),
    CONSTRAINT iam_permission_requests_target_shape_check CHECK (
        (target_kind = 'USER' AND target_user_id IS NOT NULL AND
            target_role_id IS NULL AND target_organization_id IS NULL) OR
        (target_kind = 'ROLE' AND target_user_id IS NULL AND
            target_role_id IS NOT NULL AND target_organization_id IS NULL) OR
        (target_kind = 'ORGANIZATION' AND target_user_id IS NULL AND
            target_role_id IS NULL AND target_organization_id IS NOT NULL)
    ),
    CONSTRAINT iam_permission_requests_exact_check
        CHECK (permission_match_kind = 'exact'),
    CONSTRAINT iam_permission_requests_state_check CHECK (
        state IN (
            'PENDING', 'IN_REVIEW', 'APPROVED', 'REJECTED',
            'CANCELLED', 'EXPIRED'
        )
    ),
    CONSTRAINT iam_permission_requests_revision_check CHECK (revision > 0),
    CONSTRAINT iam_permission_requests_binding_check CHECK (
        (state = 'APPROVED' AND granted_binding_id = planned_binding_id) OR
        (state <> 'APPROVED' AND granted_binding_id IS NULL)
    ),
    CONSTRAINT iam_permission_requests_timestamp_check CHECK (
        updated_at >= created_at AND
        (terminal_at IS NULL OR
            (terminal_at >= created_at AND terminal_at <= updated_at))
    ),
    CONSTRAINT iam_permission_requests_lifecycle_check CHECK (
        (state IN ('PENDING', 'IN_REVIEW') AND terminal_at IS NULL) OR
        (state IN ('APPROVED', 'REJECTED', 'CANCELLED', 'EXPIRED') AND
            terminal_at IS NOT NULL)
    )
);

CREATE UNIQUE INDEX iam_permission_requests_target_permission_uidx
    ON iam_permission_requests (
        group_id, target_kind, target_user_id, target_role_id,
        target_organization_id, permission_id
    ) NULLS NOT DISTINCT;

CREATE INDEX iam_permission_requests_group_idx
    ON iam_permission_requests (group_id, ordinal, id);
