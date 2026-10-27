-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_platform_owner_policy_guard_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE binding_kind text; binding_state text; subject_state text;
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Instance Owner policy deletion is forbidden';
    END IF;
    SELECT role.role_kind, binding.state, subject.state
    INTO binding_kind, binding_state, subject_state
    FROM platform.iam_platform_subject_roles AS binding
    JOIN platform.iam_platform_roles AS role
      ON role.environment_id = binding.environment_id AND role.id = binding.role_id
    JOIN platform.iam_platform_subjects AS subject
      ON subject.environment_id = binding.environment_id AND subject.id = binding.subject_id
    WHERE binding.environment_id = NEW.environment_id AND binding.id = NEW.owner_binding_id
    FOR SHARE OF binding, role, subject;
    IF binding_kind IS DISTINCT FROM 'instance_owner' OR
       binding_state IS DISTINCT FROM 'active' OR subject_state IS DISTINCT FROM 'active' THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Instance Owner binding is invalid';
    END IF;
    IF TG_OP = 'INSERT' AND NEW.revision IS DISTINCT FROM 1 THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Instance Owner policy revision is invalid';
    ELSIF TG_OP = 'UPDATE' AND (
        NEW.environment_id IS DISTINCT FROM OLD.environment_id OR
        NEW.owner_binding_id IS NOT DISTINCT FROM OLD.owner_binding_id OR
        NEW.created_at IS DISTINCT FROM OLD.created_at OR
        NEW.revision IS DISTINCT FROM OLD.revision + 1
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Instance Owner transfer CAS is invalid';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION platform.iam_platform_clock_guard_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Platform authorization clock deletion is forbidden';
    ELSIF TG_OP = 'INSERT' AND
          (NEW.authorization_clock IS DISTINCT FROM 1 OR NEW.revision IS DISTINCT FROM 1) THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Platform authorization clock origin is invalid';
    ELSIF TG_OP = 'UPDATE' AND (
        NEW.environment_id IS DISTINCT FROM OLD.environment_id OR
        NEW.created_at IS DISTINCT FROM OLD.created_at OR
        NEW.authorization_clock IS DISTINCT FROM OLD.authorization_clock + 1 OR
        NEW.revision IS DISTINCT FROM OLD.revision + 1
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Platform authorization clock step is invalid';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION platform.iam_platform_receipt_guard_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Platform IAM receipt is immutable';
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER iam_platform_owner_policies_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.iam_platform_owner_policies
FOR EACH ROW EXECUTE FUNCTION platform.iam_platform_owner_policy_guard_v1();
ALTER TABLE platform.iam_platform_owner_policies
ENABLE ALWAYS TRIGGER iam_platform_owner_policies_guard;

CREATE TRIGGER iam_platform_authorization_clocks_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.iam_platform_authorization_clocks
FOR EACH ROW EXECUTE FUNCTION platform.iam_platform_clock_guard_v1();
ALTER TABLE platform.iam_platform_authorization_clocks
ENABLE ALWAYS TRIGGER iam_platform_authorization_clocks_guard;

CREATE TRIGGER iam_platform_mutation_receipts_guard
BEFORE UPDATE OR DELETE ON platform.iam_platform_mutation_receipts
FOR EACH ROW EXECUTE FUNCTION platform.iam_platform_receipt_guard_v1();
ALTER TABLE platform.iam_platform_mutation_receipts
ENABLE ALWAYS TRIGGER iam_platform_mutation_receipts_guard;

REVOKE ALL ON FUNCTION platform.iam_platform_owner_policy_guard_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_platform_clock_guard_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_platform_receipt_guard_v1() FROM PUBLIC;
