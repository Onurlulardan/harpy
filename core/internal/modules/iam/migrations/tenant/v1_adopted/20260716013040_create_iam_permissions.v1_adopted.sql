-- +goose Up
CREATE TABLE iam_permissions (
    id uuid NOT NULL,
    match_kind text NOT NULL,
    resource_id uuid,
    action_id uuid,
    is_protected boolean NOT NULL,
    state text NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    retired_at timestamptz,
    CONSTRAINT iam_permissions_pkey PRIMARY KEY (id),
    CONSTRAINT iam_permissions_match_identity_key UNIQUE (id, match_kind),
    CONSTRAINT iam_permissions_identity_key UNIQUE (id, match_kind, is_protected),
    CONSTRAINT iam_permissions_policy_identity_key
        UNIQUE (id, match_kind, is_protected, state),
    CONSTRAINT iam_permissions_shape_key
        UNIQUE NULLS NOT DISTINCT (match_kind, resource_id, action_id),
    CONSTRAINT iam_permissions_resource_fk FOREIGN KEY (resource_id)
        REFERENCES iam_resources (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_permissions_action_fk FOREIGN KEY (action_id)
        REFERENCES iam_actions (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_permissions_match_kind_check CHECK (
        match_kind IN ('exact', 'resource_all_actions', 'tenant_all')
    ),
    CONSTRAINT iam_permissions_shape_check CHECK (
        (match_kind = 'exact' AND resource_id IS NOT NULL AND action_id IS NOT NULL) OR
        (match_kind = 'resource_all_actions' AND
            resource_id IS NOT NULL AND action_id IS NULL) OR
        (match_kind = 'tenant_all' AND resource_id IS NULL AND action_id IS NULL)
    ),
    CONSTRAINT iam_permissions_protected_match_check CHECK (
        match_kind = 'exact' OR is_protected
    ),
    CONSTRAINT iam_permissions_state_check CHECK (state IN ('active', 'retired')),
    CONSTRAINT iam_permissions_revision_check CHECK (revision > 0),
    CONSTRAINT iam_permissions_timestamp_order_check CHECK (updated_at >= created_at),
    CONSTRAINT iam_permissions_retirement_check CHECK (
        (state = 'active' AND retired_at IS NULL) OR
        (state = 'retired' AND retired_at IS NOT NULL AND retired_at >= updated_at)
    )
);

-- +goose StatementBegin
CREATE FUNCTION iam_guard_active_permission_catalog_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    resource_active boolean := true;
    action_active boolean := true;
BEGIN
    IF TG_OP NOT IN ('INSERT', 'UPDATE') THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM permission catalog guard invocation is invalid';
    END IF;
    IF NEW.state IS DISTINCT FROM 'active' THEN
        RETURN NEW;
    END IF;
    IF NEW.resource_id IS NOT NULL THEN
        SELECT resource.state = 'active' INTO resource_active
        FROM iam_resources AS resource WHERE resource.id = NEW.resource_id;
    END IF;
    IF NEW.action_id IS NOT NULL THEN
        SELECT action.state = 'active' INTO action_active
        FROM iam_actions AS action WHERE action.id = NEW.action_id;
    END IF;
    IF resource_active IS DISTINCT FROM true OR action_active IS DISTINCT FROM true THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_permissions_active_catalog_guard',
            MESSAGE = 'Active permission requires active catalog entries';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_guard_active_permission_catalog_v1() FROM PUBLIC;

CREATE TRIGGER iam_permissions_revision_guard
BEFORE UPDATE ON iam_permissions
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_permissions ENABLE ALWAYS TRIGGER iam_permissions_revision_guard;

CREATE TRIGGER iam_permissions_immutable_guard
BEFORE UPDATE ON iam_permissions
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'match_kind', 'resource_id', 'action_id', 'is_protected', 'created_at'
);
ALTER TABLE iam_permissions ENABLE ALWAYS TRIGGER iam_permissions_immutable_guard;

CREATE TRIGGER iam_permissions_state_guard
BEFORE UPDATE ON iam_permissions
FOR EACH ROW EXECUTE FUNCTION iam_guard_state_transition_v1('active>retired');
ALTER TABLE iam_permissions ENABLE ALWAYS TRIGGER iam_permissions_state_guard;

CREATE TRIGGER iam_permissions_terminal_guard
BEFORE UPDATE ON iam_permissions
FOR EACH ROW EXECUTE FUNCTION iam_guard_rbac_terminal_v1();
ALTER TABLE iam_permissions ENABLE ALWAYS TRIGGER iam_permissions_terminal_guard;

CREATE TRIGGER iam_permissions_catalog_guard
BEFORE INSERT OR UPDATE ON iam_permissions
FOR EACH ROW EXECUTE FUNCTION iam_guard_active_permission_catalog_v1();
ALTER TABLE iam_permissions ENABLE ALWAYS TRIGGER iam_permissions_catalog_guard;

CREATE TRIGGER iam_permissions_removal_guard
BEFORE DELETE OR TRUNCATE ON iam_permissions
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_permissions ENABLE ALWAYS TRIGGER iam_permissions_removal_guard;
