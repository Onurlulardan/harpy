-- +goose Up
ALTER TABLE iam_user_role_assignments
    ADD CONSTRAINT iam_user_role_assignments_portal_key
    UNIQUE (id, user_id, role_id);

CREATE TABLE iam_portal_role_assignments (
    binding_id uuid NOT NULL,
    user_id uuid NOT NULL,
    assignment_id uuid NOT NULL,
    role_id uuid NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT iam_portal_role_assignments_pkey
        PRIMARY KEY (binding_id, role_id),
    CONSTRAINT iam_portal_role_assignments_assignment_key
        UNIQUE (assignment_id),
    CONSTRAINT iam_portal_role_assignments_binding_fk
        FOREIGN KEY (binding_id, user_id)
        REFERENCES iam_portal_subject_bindings (id, user_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_portal_role_assignments_assignment_fk
        FOREIGN KEY (assignment_id, user_id, role_id)
        REFERENCES iam_user_role_assignments (id, user_id, role_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE TABLE iam_portal_console_grants (
    user_id uuid NOT NULL,
    access_mode_id uuid NOT NULL,
    access_mode text NOT NULL,
    assignment_id uuid NOT NULL,
    role_id uuid NOT NULL,
    source text NOT NULL,
    state text NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    revoked_at timestamptz,
    CONSTRAINT iam_portal_console_grants_pkey PRIMARY KEY (user_id),
    CONSTRAINT iam_portal_console_grants_access_fk
        FOREIGN KEY (access_mode_id, user_id, access_mode)
        REFERENCES iam_user_access_modes (id, user_id, mode)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_portal_console_grants_assignment_fk
        FOREIGN KEY (assignment_id, user_id, role_id)
        REFERENCES iam_user_role_assignments (id, user_id, role_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_portal_console_grants_mode_check
        CHECK (access_mode = 'operator_console'),
    CONSTRAINT iam_portal_console_grants_source_check
        CHECK (source IN ('portal_grant', 'preexisting_operator')),
    CONSTRAINT iam_portal_console_grants_state_check
        CHECK (state IN ('active', 'revoked')),
    CONSTRAINT iam_portal_console_grants_revision_check CHECK (revision > 0),
    CONSTRAINT iam_portal_console_grants_timestamp_check
        CHECK (updated_at >= created_at),
    CONSTRAINT iam_portal_console_grants_revocation_check CHECK (
        (state = 'active' AND revoked_at IS NULL) OR
        (state = 'revoked' AND revoked_at IS NOT NULL AND revoked_at >= updated_at)
    )
);

CREATE TRIGGER iam_portal_console_grants_revision_guard
BEFORE UPDATE ON iam_portal_console_grants
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_portal_console_grants
    ENABLE ALWAYS TRIGGER iam_portal_console_grants_revision_guard;

CREATE TRIGGER iam_portal_console_grants_immutable_guard
BEFORE UPDATE ON iam_portal_console_grants
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'user_id', 'access_mode_id', 'access_mode', 'assignment_id',
    'role_id', 'source', 'created_at'
);
ALTER TABLE iam_portal_console_grants
    ENABLE ALWAYS TRIGGER iam_portal_console_grants_immutable_guard;

ALTER TABLE iam_identity_change_audit
    DROP CONSTRAINT iam_identity_change_audit_command_check,
    ADD CONSTRAINT iam_identity_change_audit_command_check CHECK (
        command_key IN (
            'iam.identity.user.create', 'iam.identity.user.update',
            'iam.identity.user.disable', 'iam.identity.user.delete',
            'iam.identity.role.create', 'iam.identity.role.update',
            'iam.identity.role.retire', 'iam.identity.organization.create',
            'iam.identity.organization.update', 'iam.identity.organization.disable',
            'iam.identity.organization.delete', 'iam.portal_user.create',
            'iam.portal_user.status', 'iam.portal_user.console.grant',
            'iam.portal_user.console.revoke'
        )
    );

