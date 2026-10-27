-- +goose Up
CREATE TABLE iam_rbac_assignment_members (
    group_id uuid NOT NULL,
    group_revision bigint NOT NULL,
    ordinal integer NOT NULL,
    action_id uuid NOT NULL,
    permission_id uuid NOT NULL,
    permission_match_kind text NOT NULL DEFAULT 'exact',
    binding_id uuid NOT NULL,
    state text NOT NULL,
    created_at timestamptz NOT NULL,
    revoked_at timestamptz,
    CONSTRAINT iam_rbac_assignment_members_pkey
        PRIMARY KEY (group_id, group_revision, ordinal),
    CONSTRAINT iam_rbac_assignment_members_group_fk FOREIGN KEY (group_id)
        REFERENCES iam_rbac_assignment_groups (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_rbac_assignment_members_action_fk FOREIGN KEY (action_id)
        REFERENCES iam_actions (id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_rbac_assignment_members_permission_fk
        FOREIGN KEY (permission_id, permission_match_kind)
        REFERENCES iam_permissions (id, match_kind)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_rbac_assignment_members_revision_check CHECK (group_revision > 0),
    CONSTRAINT iam_rbac_assignment_members_ordinal_check CHECK (ordinal > 0),
    CONSTRAINT iam_rbac_assignment_members_exact_check
        CHECK (permission_match_kind = 'exact'),
    CONSTRAINT iam_rbac_assignment_members_state_check
        CHECK (state IN ('active', 'revoked')),
    CONSTRAINT iam_rbac_assignment_members_lifecycle_check CHECK (
        (state = 'active' AND revoked_at IS NULL) OR
        (state = 'revoked' AND revoked_at IS NOT NULL AND revoked_at >= created_at)
    ),
    CONSTRAINT iam_rbac_assignment_members_action_key
        UNIQUE (group_id, group_revision, action_id),
    CONSTRAINT iam_rbac_assignment_members_binding_key
        UNIQUE (group_id, group_revision, binding_id)
);

CREATE INDEX iam_rbac_assignment_members_permission_idx
    ON iam_rbac_assignment_members (permission_id, state);

CREATE TRIGGER iam_rbac_assignment_members_immutable_guard
BEFORE UPDATE ON iam_rbac_assignment_members
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'group_id', 'group_revision', 'ordinal', 'action_id',
    'permission_id', 'permission_match_kind', 'binding_id', 'created_at'
);
ALTER TABLE iam_rbac_assignment_members
    ENABLE ALWAYS TRIGGER iam_rbac_assignment_members_immutable_guard;

CREATE TRIGGER iam_rbac_assignment_members_state_guard
BEFORE UPDATE ON iam_rbac_assignment_members
FOR EACH ROW EXECUTE FUNCTION iam_guard_state_transition_v1('active>revoked');
ALTER TABLE iam_rbac_assignment_members
    ENABLE ALWAYS TRIGGER iam_rbac_assignment_members_state_guard;

CREATE TRIGGER iam_rbac_assignment_members_removal_guard
BEFORE DELETE OR TRUNCATE ON iam_rbac_assignment_members
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_rbac_assignment_members
    ENABLE ALWAYS TRIGGER iam_rbac_assignment_members_removal_guard;
