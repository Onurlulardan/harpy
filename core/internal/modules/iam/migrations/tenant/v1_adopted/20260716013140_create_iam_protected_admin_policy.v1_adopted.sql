-- +goose Up
CREATE TABLE iam_protected_admin_policies (
    singleton boolean NOT NULL,
    role_id uuid NOT NULL,
    role_is_protected boolean NOT NULL,
    role_status text NOT NULL,
    role_scope_kind text NOT NULL,
    tenant_all_permission_id uuid NOT NULL,
    permission_match_kind text NOT NULL,
    permission_is_protected boolean NOT NULL,
    permission_state text NOT NULL,
    role_permission_binding_id uuid NOT NULL,
    role_binding_state text NOT NULL,
    minimum_active_assignments integer NOT NULL,
    policy_state text NOT NULL,
    authorization_epoch bigint NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT iam_protected_admin_policies_pkey PRIMARY KEY (singleton),
    CONSTRAINT iam_protected_admin_policies_role_fk
        FOREIGN KEY (role_id, role_is_protected, role_status, role_scope_kind)
        REFERENCES iam_roles (id, is_protected, status, scope_kind)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_protected_admin_policies_permission_fk
        FOREIGN KEY (
            tenant_all_permission_id, permission_match_kind,
            permission_is_protected, permission_state
        ) REFERENCES iam_permissions (id, match_kind, is_protected, state)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_protected_admin_policies_binding_fk
        FOREIGN KEY (
            role_permission_binding_id, role_id,
            tenant_all_permission_id, role_binding_state
        ) REFERENCES iam_role_permission_bindings (
            id, role_id, permission_id, state
        ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_protected_admin_policies_singleton_check CHECK (singleton),
    CONSTRAINT iam_protected_admin_policies_role_check CHECK (
        role_is_protected AND role_status = 'active' AND role_scope_kind = 'tenant'
    ),
    CONSTRAINT iam_protected_admin_policies_permission_check CHECK (
        permission_match_kind = 'tenant_all' AND
        permission_is_protected AND permission_state = 'active'
    ),
    CONSTRAINT iam_protected_admin_policies_binding_check
        CHECK (role_binding_state = 'active'),
    CONSTRAINT iam_protected_admin_policies_minimum_check
        CHECK (minimum_active_assignments >= 1),
    CONSTRAINT iam_protected_admin_policies_state_check
        CHECK (policy_state = 'active'),
    CONSTRAINT iam_protected_admin_policies_epoch_check
        CHECK (authorization_epoch > 0),
    CONSTRAINT iam_protected_admin_policies_revision_check CHECK (revision > 0),
    CONSTRAINT iam_protected_admin_policies_timestamp_order_check
        CHECK (updated_at >= created_at)
);

-- +goose StatementBegin
CREATE FUNCTION iam_guard_protected_admin_policy_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    active_assignments bigint;
BEGIN
    IF TG_OP NOT IN ('INSERT', 'UPDATE') THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM protected admin policy guard invocation is invalid';
    END IF;
    IF TG_OP = 'UPDATE' AND NEW.authorization_epoch <= OLD.authorization_epoch THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_protected_admin_policies_epoch_guard',
            MESSAGE = 'Protected admin policy epoch must increase';
    END IF;
    SELECT pg_catalog.count(*) INTO active_assignments
    FROM iam_user_role_assignments AS assignment
    JOIN iam_users AS subject ON subject.id = assignment.user_id
    WHERE assignment.role_id = NEW.role_id
      AND assignment.state = 'active' AND subject.status = 'active';
    IF active_assignments < NEW.minimum_active_assignments THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_protected_admin_policies_minimum_guard',
            MESSAGE = 'Protected admin policy requires active assignments';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_guard_protected_admin_policy_v1() FROM PUBLIC;

CREATE TRIGGER iam_protected_admin_policies_revision_guard
BEFORE UPDATE ON iam_protected_admin_policies
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_protected_admin_policies
    ENABLE ALWAYS TRIGGER iam_protected_admin_policies_revision_guard;

CREATE TRIGGER iam_protected_admin_policies_immutable_guard
BEFORE UPDATE ON iam_protected_admin_policies
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'singleton', 'created_at'
);
ALTER TABLE iam_protected_admin_policies
    ENABLE ALWAYS TRIGGER iam_protected_admin_policies_immutable_guard;

CREATE TRIGGER iam_protected_admin_policies_policy_guard
BEFORE INSERT OR UPDATE ON iam_protected_admin_policies
FOR EACH ROW EXECUTE FUNCTION iam_guard_protected_admin_policy_v1();
ALTER TABLE iam_protected_admin_policies
    ENABLE ALWAYS TRIGGER iam_protected_admin_policies_policy_guard;

CREATE TRIGGER iam_protected_admin_policies_removal_guard
BEFORE DELETE OR TRUNCATE ON iam_protected_admin_policies
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_protected_admin_policies
    ENABLE ALWAYS TRIGGER iam_protected_admin_policies_removal_guard;
