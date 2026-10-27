-- +goose Up
CREATE TABLE iam_rbac_assignment_groups (
    id uuid NOT NULL,
    resource_id uuid NOT NULL,
    target_kind text COLLATE "C" NOT NULL,
    target_id uuid NOT NULL,
    state text NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    retired_at timestamptz,
    CONSTRAINT iam_rbac_assignment_groups_pkey PRIMARY KEY (id),
    CONSTRAINT iam_rbac_assignment_groups_resource_fk FOREIGN KEY (resource_id)
        REFERENCES iam_resources (id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_rbac_assignment_groups_target_check CHECK (
        target_kind IN ('USER', 'ROLE', 'ORGANIZATION')
    ),
    CONSTRAINT iam_rbac_assignment_groups_state_check CHECK (
        state IN ('active', 'retired')
    ),
    CONSTRAINT iam_rbac_assignment_groups_revision_check CHECK (revision > 0),
    CONSTRAINT iam_rbac_assignment_groups_timestamp_check CHECK (
        updated_at >= created_at AND (
            (state = 'active' AND retired_at IS NULL) OR
            (state = 'retired' AND retired_at IS NOT NULL AND retired_at >= updated_at)
        )
    )
);

CREATE UNIQUE INDEX iam_rbac_assignment_groups_active_subject_resource_uidx
    ON iam_rbac_assignment_groups (target_kind, target_id, resource_id)
    WHERE state = 'active';

CREATE TRIGGER iam_rbac_assignment_groups_revision_guard
BEFORE UPDATE ON iam_rbac_assignment_groups
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_rbac_assignment_groups
    ENABLE ALWAYS TRIGGER iam_rbac_assignment_groups_revision_guard;

CREATE TRIGGER iam_rbac_assignment_groups_immutable_guard
BEFORE UPDATE ON iam_rbac_assignment_groups
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1('id', 'created_at');
ALTER TABLE iam_rbac_assignment_groups
    ENABLE ALWAYS TRIGGER iam_rbac_assignment_groups_immutable_guard;

CREATE TRIGGER iam_rbac_assignment_groups_state_guard
BEFORE UPDATE ON iam_rbac_assignment_groups
FOR EACH ROW EXECUTE FUNCTION iam_guard_state_transition_v1('active>retired');
ALTER TABLE iam_rbac_assignment_groups
    ENABLE ALWAYS TRIGGER iam_rbac_assignment_groups_state_guard;

CREATE TRIGGER iam_rbac_assignment_groups_terminal_guard
BEFORE UPDATE ON iam_rbac_assignment_groups
FOR EACH ROW EXECUTE FUNCTION iam_guard_rbac_terminal_v1();
ALTER TABLE iam_rbac_assignment_groups
    ENABLE ALWAYS TRIGGER iam_rbac_assignment_groups_terminal_guard;

CREATE TRIGGER iam_rbac_assignment_groups_removal_guard
BEFORE DELETE OR TRUNCATE ON iam_rbac_assignment_groups
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_rbac_assignment_groups
    ENABLE ALWAYS TRIGGER iam_rbac_assignment_groups_removal_guard;
