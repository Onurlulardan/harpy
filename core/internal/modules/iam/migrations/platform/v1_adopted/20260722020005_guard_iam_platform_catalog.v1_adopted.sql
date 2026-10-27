-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_platform_permission_guard_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE exact_key text;
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Platform permission deletion is forbidden';
    END IF;
    IF NEW.match_kind = 'exact' THEN
        SELECT resource.resource_key || '.' || action.action_key INTO exact_key
        FROM platform.iam_platform_resources AS resource
        JOIN platform.iam_platform_actions AS action ON action.id = NEW.action_id
        WHERE resource.id = NEW.resource_id AND resource.state = 'active'
          AND action.state = 'active' FOR SHARE OF resource, action;
        IF exact_key IS DISTINCT FROM NEW.permission_key THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'Platform permission key does not match its exact catalog pair';
        END IF;
    ELSIF NEW.permission_key IS DISTINCT FROM 'environment_all' OR
          NEW.delegable OR NOT NEW.owner_only OR NEW.deployment_mode <> 'both' THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'Reserved environment permission shape is invalid';
    END IF;
    IF TG_OP = 'UPDATE' AND (
        NEW.id IS DISTINCT FROM OLD.id OR NEW.permission_key IS DISTINCT FROM OLD.permission_key OR
        NEW.match_kind IS DISTINCT FROM OLD.match_kind OR NEW.resource_id IS DISTINCT FROM OLD.resource_id OR
        NEW.action_id IS DISTINCT FROM OLD.action_id OR
        NEW.deployment_mode IS DISTINCT FROM OLD.deployment_mode OR
        NEW.delegable IS DISTINCT FROM OLD.delegable OR
        NEW.step_up_required IS DISTINCT FROM OLD.step_up_required OR
        NEW.owner_only IS DISTINCT FROM OLD.owner_only OR NEW.created_at IS DISTINCT FROM OLD.created_at OR
        OLD.state IS DISTINCT FROM 'active' OR NEW.state IS DISTINCT FROM 'retired' OR
        NEW.revision IS DISTINCT FROM OLD.revision + 1
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Platform permission transition is invalid';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION platform.iam_platform_role_guard_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' OR (TG_OP = 'UPDATE' AND OLD.role_kind = 'instance_owner') THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Reserved platform role mutation is forbidden';
    END IF;
    IF TG_OP = 'UPDATE' AND (
        NEW.id IS DISTINCT FROM OLD.id OR NEW.environment_id IS DISTINCT FROM OLD.environment_id OR
        NEW.role_kind IS DISTINCT FROM OLD.role_kind OR NEW.role_key IS DISTINCT FROM OLD.role_key OR
        NEW.created_at IS DISTINCT FROM OLD.created_at OR OLD.state IS DISTINCT FROM 'active' OR
        NEW.state IS DISTINCT FROM 'retired' OR NEW.revision IS DISTINCT FROM OLD.revision + 1
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Platform role transition is invalid';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER iam_platform_permissions_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.iam_platform_permissions
FOR EACH ROW EXECUTE FUNCTION platform.iam_platform_permission_guard_v1();
ALTER TABLE platform.iam_platform_permissions
ENABLE ALWAYS TRIGGER iam_platform_permissions_guard;

CREATE TRIGGER iam_platform_roles_guard
BEFORE UPDATE OR DELETE ON platform.iam_platform_roles
FOR EACH ROW EXECUTE FUNCTION platform.iam_platform_role_guard_v1();
ALTER TABLE platform.iam_platform_roles ENABLE ALWAYS TRIGGER iam_platform_roles_guard;

REVOKE ALL ON FUNCTION platform.iam_platform_permission_guard_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_platform_role_guard_v1() FROM PUBLIC;
