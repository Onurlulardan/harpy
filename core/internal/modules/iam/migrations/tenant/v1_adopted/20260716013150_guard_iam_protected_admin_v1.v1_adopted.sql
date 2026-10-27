-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_guard_role_wildcard_policy_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    policy_exists boolean;
BEGIN
    IF TG_OP NOT IN ('INSERT', 'UPDATE') THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM role wildcard guard invocation is invalid';
    END IF;
    IF NEW.state IS DISTINCT FROM 'active' OR
       NEW.permission_match_kind IS NOT DISTINCT FROM 'exact' THEN
        RETURN NEW;
    END IF;
    SELECT true INTO policy_exists
    FROM iam_protected_admin_policies AS policy
    WHERE policy.role_id = NEW.role_id AND policy.policy_state = 'active';
    IF policy_exists IS DISTINCT FROM true THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_role_permission_bindings_wildcard_policy_guard',
            MESSAGE = 'Wildcard binding requires the protected admin policy role';
    END IF;
    RETURN NEW;
END;
$function$;
REVOKE ALL ON FUNCTION iam_guard_role_wildcard_policy_v1() FROM PUBLIC;

CREATE FUNCTION iam_guard_policy_wildcard_move_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM protected policy move guard invocation is invalid';
    END IF;
    IF NEW.role_id IS NOT DISTINCT FROM OLD.role_id THEN
        RETURN NEW;
    END IF;
    IF EXISTS (
        SELECT 1 FROM iam_role_permission_bindings AS binding
        WHERE binding.role_id = OLD.role_id AND binding.state = 'active'
          AND binding.permission_match_kind <> 'exact'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_protected_admin_policies_wildcard_move_guard',
            MESSAGE = 'Old protected role retains an active wildcard binding';
    END IF;
    RETURN NEW;
END;
$function$;
REVOKE ALL ON FUNCTION iam_guard_policy_wildcard_move_v1() FROM PUBLIC;

CREATE FUNCTION iam_guard_protected_admin_minimum_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    protected_role uuid;
    required_count integer;
    remaining_count bigint;
BEGIN
    IF TG_TABLE_NAME = 'iam_user_role_assignments' THEN
        IF OLD.state IS DISTINCT FROM 'active' OR
           (TG_OP = 'UPDATE' AND NEW.state IS NOT DISTINCT FROM 'active') THEN
            IF TG_OP = 'DELETE' THEN
                RETURN OLD;
            END IF;
            RETURN NEW;
        END IF;
        SELECT policy.role_id, policy.minimum_active_assignments
        INTO protected_role, required_count
        FROM iam_protected_admin_policies AS policy
        WHERE policy.role_id = OLD.role_id FOR UPDATE;
        IF protected_role IS NULL THEN
            IF TG_OP = 'DELETE' THEN
                RETURN OLD;
            END IF;
            RETURN NEW;
        END IF;
        SELECT pg_catalog.count(*) INTO remaining_count
        FROM iam_user_role_assignments AS assignment
        JOIN iam_users AS subject ON subject.id = assignment.user_id
        WHERE assignment.role_id = protected_role AND assignment.state = 'active'
          AND subject.status = 'active' AND assignment.id <> OLD.id;
    ELSIF TG_TABLE_NAME = 'iam_users' AND TG_OP = 'UPDATE' THEN
        IF OLD.status IS DISTINCT FROM 'active' OR NEW.status = 'active' THEN
            RETURN NEW;
        END IF;
        SELECT policy.role_id, policy.minimum_active_assignments
        INTO protected_role, required_count
        FROM iam_protected_admin_policies AS policy
        JOIN iam_user_role_assignments AS assignment
          ON assignment.role_id = policy.role_id
        WHERE assignment.user_id = OLD.id AND assignment.state = 'active'
        FOR UPDATE OF policy;
        IF protected_role IS NULL THEN
            RETURN NEW;
        END IF;
        SELECT pg_catalog.count(*) INTO remaining_count
        FROM iam_user_role_assignments AS assignment
        JOIN iam_users AS subject ON subject.id = assignment.user_id
        WHERE assignment.role_id = protected_role AND assignment.state = 'active'
          AND subject.status = 'active' AND subject.id <> OLD.id;
    ELSE
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM protected admin minimum guard invocation is invalid';
    END IF;
    IF remaining_count < required_count THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_protected_admin_minimum_guard',
            MESSAGE = 'Protected admin assignment minimum would be violated';
    END IF;
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$function$;
REVOKE ALL ON FUNCTION iam_guard_protected_admin_minimum_v1() FROM PUBLIC;
-- +goose StatementEnd

CREATE CONSTRAINT TRIGGER iam_role_permission_bindings_wildcard_policy_guard
AFTER INSERT OR UPDATE ON iam_role_permission_bindings
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION iam_guard_role_wildcard_policy_v1();
ALTER TABLE iam_role_permission_bindings
    ENABLE ALWAYS TRIGGER iam_role_permission_bindings_wildcard_policy_guard;

CREATE CONSTRAINT TRIGGER iam_protected_admin_policies_wildcard_move_guard
AFTER UPDATE OF role_id ON iam_protected_admin_policies
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION iam_guard_policy_wildcard_move_v1();
ALTER TABLE iam_protected_admin_policies
    ENABLE ALWAYS TRIGGER iam_protected_admin_policies_wildcard_move_guard;

CREATE TRIGGER iam_user_role_assignments_protected_minimum_guard
BEFORE UPDATE OR DELETE ON iam_user_role_assignments
FOR EACH ROW EXECUTE FUNCTION iam_guard_protected_admin_minimum_v1();
ALTER TABLE iam_user_role_assignments
    ENABLE ALWAYS TRIGGER iam_user_role_assignments_protected_minimum_guard;

CREATE TRIGGER iam_user_role_assignments_protected_truncate_guard
BEFORE TRUNCATE ON iam_user_role_assignments
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_user_role_assignments
    ENABLE ALWAYS TRIGGER iam_user_role_assignments_protected_truncate_guard;

CREATE TRIGGER iam_users_protected_minimum_guard
BEFORE UPDATE OF status ON iam_users
FOR EACH ROW EXECUTE FUNCTION iam_guard_protected_admin_minimum_v1();
ALTER TABLE iam_users ENABLE ALWAYS TRIGGER iam_users_protected_minimum_guard;
