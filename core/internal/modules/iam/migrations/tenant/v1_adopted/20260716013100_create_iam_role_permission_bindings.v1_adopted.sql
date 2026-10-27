-- +goose Up
ALTER TABLE iam_roles ADD CONSTRAINT iam_roles_protected_identity_key
    UNIQUE (id, is_protected);
ALTER TABLE iam_roles ADD CONSTRAINT iam_roles_policy_identity_key
    UNIQUE (id, is_protected, status, scope_kind);

CREATE TABLE iam_role_permission_bindings (
    id uuid NOT NULL,
    role_id uuid NOT NULL,
    role_is_protected boolean NOT NULL,
    permission_id uuid NOT NULL,
    permission_match_kind text NOT NULL,
    state text NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    activated_at timestamptz NOT NULL,
    revoked_at timestamptz,
    CONSTRAINT iam_role_permission_bindings_pkey PRIMARY KEY (id),
    CONSTRAINT iam_role_permission_bindings_subject_permission_key
        UNIQUE (role_id, permission_id),
    CONSTRAINT iam_role_permission_bindings_policy_identity_key
        UNIQUE (id, role_id, permission_id, state),
    CONSTRAINT iam_role_permission_bindings_role_fk
        FOREIGN KEY (role_id, role_is_protected)
        REFERENCES iam_roles (id, is_protected)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_role_permission_bindings_permission_fk
        FOREIGN KEY (permission_id, permission_match_kind)
        REFERENCES iam_permissions (id, match_kind)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_role_permission_bindings_wildcard_check CHECK (
        permission_match_kind = 'exact' OR role_is_protected
    ),
    CONSTRAINT iam_role_permission_bindings_state_check
        CHECK (state IN ('active', 'revoked')),
    CONSTRAINT iam_role_permission_bindings_revision_check CHECK (revision > 0),
    CONSTRAINT iam_role_permission_bindings_timestamp_order_check CHECK (
        updated_at >= created_at AND activated_at >= created_at
    ),
    CONSTRAINT iam_role_permission_bindings_lifecycle_check CHECK (
        (state = 'active' AND revoked_at IS NULL AND activated_at >= updated_at) OR
        (state = 'revoked' AND revoked_at IS NOT NULL AND revoked_at >= updated_at)
    )
);

CREATE INDEX iam_role_permission_bindings_permission_idx
    ON iam_role_permission_bindings (permission_id, state);

-- +goose StatementBegin
CREATE FUNCTION iam_guard_active_role_binding_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    eligible boolean;
BEGIN
    IF TG_OP NOT IN ('INSERT', 'UPDATE') THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM role binding guard invocation is invalid';
    END IF;
    IF NEW.state IS DISTINCT FROM 'active' THEN
        RETURN NEW;
    END IF;
    SELECT true INTO eligible
    FROM iam_roles AS subject
    JOIN iam_permissions AS permission ON permission.id = NEW.permission_id
    LEFT JOIN iam_resources AS resource ON resource.id = permission.resource_id
    LEFT JOIN iam_actions AS action ON action.id = permission.action_id
    WHERE subject.id = NEW.role_id AND subject.status = 'active'
      AND subject.is_protected = NEW.role_is_protected
      AND permission.match_kind = NEW.permission_match_kind
      AND permission.state = 'active'
      AND (resource.id IS NULL OR resource.state = 'active')
      AND (action.id IS NULL OR action.state = 'active');
    IF eligible IS DISTINCT FROM true THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_role_permission_bindings_active_guard',
            MESSAGE = 'Active role binding requires active subject and permission';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_guard_active_role_binding_v1() FROM PUBLIC;

CREATE TRIGGER iam_role_permission_bindings_revision_guard
BEFORE UPDATE ON iam_role_permission_bindings
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_role_permission_bindings
    ENABLE ALWAYS TRIGGER iam_role_permission_bindings_revision_guard;

CREATE TRIGGER iam_role_permission_bindings_immutable_guard
BEFORE UPDATE ON iam_role_permission_bindings
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'role_id', 'role_is_protected', 'permission_id',
    'permission_match_kind', 'created_at'
);
ALTER TABLE iam_role_permission_bindings
    ENABLE ALWAYS TRIGGER iam_role_permission_bindings_immutable_guard;

CREATE TRIGGER iam_role_permission_bindings_state_guard
BEFORE UPDATE ON iam_role_permission_bindings
FOR EACH ROW EXECUTE FUNCTION iam_guard_state_transition_v1(
    'active>revoked', 'revoked>active'
);
ALTER TABLE iam_role_permission_bindings
    ENABLE ALWAYS TRIGGER iam_role_permission_bindings_state_guard;

CREATE TRIGGER iam_role_permission_bindings_active_guard
BEFORE INSERT OR UPDATE ON iam_role_permission_bindings
FOR EACH ROW EXECUTE FUNCTION iam_guard_active_role_binding_v1();
ALTER TABLE iam_role_permission_bindings
    ENABLE ALWAYS TRIGGER iam_role_permission_bindings_active_guard;

CREATE TRIGGER iam_role_permission_bindings_removal_guard
BEFORE DELETE OR TRUNCATE ON iam_role_permission_bindings
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_role_permission_bindings
    ENABLE ALWAYS TRIGGER iam_role_permission_bindings_removal_guard;
