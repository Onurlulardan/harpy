-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_platform_subject_guard_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE binding_purpose text;
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Platform subject deletion is forbidden';
    END IF;
    SELECT binding.purpose INTO binding_purpose
    FROM platform.tenancy_tenant_bindings AS binding
    WHERE binding.environment_id = NEW.environment_id
      AND binding.tenant_id = NEW.primary_tenant_id
    FOR SHARE;
    IF binding_purpose NOT IN ('onpremise_primary', 'saas_control') THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Platform subject must belong to the immutable primary tenant';
    END IF;
    IF TG_OP = 'UPDATE' AND (
        NEW.id IS DISTINCT FROM OLD.id OR
        NEW.environment_id IS DISTINCT FROM OLD.environment_id OR
        NEW.primary_tenant_id IS DISTINCT FROM OLD.primary_tenant_id OR
        NEW.user_id IS DISTINCT FROM OLD.user_id OR
        NEW.created_at IS DISTINCT FROM OLD.created_at OR
        OLD.state IS DISTINCT FROM 'active' OR NEW.state IS DISTINCT FROM 'retired' OR
        NEW.revision IS DISTINCT FROM OLD.revision + 1
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Platform subject transition is invalid';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION platform.iam_platform_role_permission_guard_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE role_kind text; permission_match text; permission_delegable boolean;
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Platform role permission deletion is forbidden';
    END IF;
    SELECT role.role_kind, permission.match_kind, permission.delegable
    INTO role_kind, permission_match, permission_delegable
    FROM platform.iam_platform_roles AS role
    JOIN platform.iam_platform_permissions AS permission ON permission.id = NEW.permission_id
    WHERE role.environment_id = NEW.environment_id AND role.id = NEW.role_id
      AND role.state = 'active' AND permission.state = 'active'
    FOR SHARE OF role, permission;
    IF (role_kind = 'instance_owner' AND permission_match <> 'environment_all') OR
       (role_kind = 'platform_operator' AND
        (permission_match <> 'exact' OR permission_delegable IS DISTINCT FROM true)) OR
       role_kind IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Platform role permission is reserved';
    END IF;
    IF TG_OP = 'UPDATE' AND (
        NEW.id IS DISTINCT FROM OLD.id OR NEW.environment_id IS DISTINCT FROM OLD.environment_id OR
        NEW.role_id IS DISTINCT FROM OLD.role_id OR NEW.permission_id IS DISTINCT FROM OLD.permission_id OR
        NEW.created_at IS DISTINCT FROM OLD.created_at OR OLD.state IS DISTINCT FROM 'active' OR
        NEW.state IS DISTINCT FROM 'revoked' OR NEW.revision IS DISTINCT FROM OLD.revision + 1
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Platform role permission transition is invalid';
    END IF;
    IF TG_OP = 'UPDATE' AND role_kind = 'instance_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Owner environment authority is immutable';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION platform.iam_platform_subject_role_guard_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE role_kind text; current_owner uuid;
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Platform subject role deletion is forbidden';
    END IF;
    SELECT role.role_kind INTO role_kind FROM platform.iam_platform_roles AS role
    WHERE role.environment_id = NEW.environment_id AND role.id = NEW.role_id
      AND role.state = 'active' FOR SHARE;
    IF role_kind IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Platform role is not active';
    END IF;
    IF TG_OP = 'UPDATE' THEN
        SELECT policy.owner_binding_id INTO current_owner
        FROM platform.iam_platform_owner_policies AS policy
        WHERE policy.environment_id = OLD.environment_id FOR SHARE;
        IF OLD.id = current_owner OR
           NEW.id IS DISTINCT FROM OLD.id OR NEW.environment_id IS DISTINCT FROM OLD.environment_id OR
           NEW.subject_id IS DISTINCT FROM OLD.subject_id OR NEW.role_id IS DISTINCT FROM OLD.role_id OR
           NEW.created_at IS DISTINCT FROM OLD.created_at OR OLD.state IS DISTINCT FROM 'active' OR
           NEW.state IS DISTINCT FROM 'revoked' OR NEW.revision IS DISTINCT FROM OLD.revision + 1 THEN
            RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Platform subject role transition is invalid';
        END IF;
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER iam_platform_subjects_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.iam_platform_subjects
FOR EACH ROW EXECUTE FUNCTION platform.iam_platform_subject_guard_v1();
ALTER TABLE platform.iam_platform_subjects ENABLE ALWAYS TRIGGER iam_platform_subjects_guard;

CREATE TRIGGER iam_platform_role_permissions_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.iam_platform_role_permissions
FOR EACH ROW EXECUTE FUNCTION platform.iam_platform_role_permission_guard_v1();
ALTER TABLE platform.iam_platform_role_permissions ENABLE ALWAYS TRIGGER iam_platform_role_permissions_guard;

CREATE TRIGGER iam_platform_subject_roles_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.iam_platform_subject_roles
FOR EACH ROW EXECUTE FUNCTION platform.iam_platform_subject_role_guard_v1();
ALTER TABLE platform.iam_platform_subject_roles ENABLE ALWAYS TRIGGER iam_platform_subject_roles_guard;

REVOKE ALL ON FUNCTION platform.iam_platform_subject_guard_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_platform_role_permission_guard_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_platform_subject_role_guard_v1() FROM PUBLIC;
