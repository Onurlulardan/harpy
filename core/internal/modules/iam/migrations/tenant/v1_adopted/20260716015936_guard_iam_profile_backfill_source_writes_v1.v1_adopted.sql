-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_guard_profile_backfill_source_write_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_target_schema_id uuid;
BEGIN
    IF TG_OP='INSERT' THEN
        IF EXISTS (
            SELECT 1 FROM iam_profile_backfill_plans AS plan
            WHERE plan.source_schema_id=NEW.profile_schema_version_id
              AND plan.resource_kind=TG_ARGV[0]
              AND plan.state IN ('planned','running','verified')
        ) THEN
            RAISE EXCEPTION USING ERRCODE='40001',
                MESSAGE='IAM profile source is fenced for backfill';
        END IF;
        RETURN NEW;
    END IF;
    IF NEW.profile_schema_version_id IS NOT DISTINCT FROM OLD.profile_schema_version_id AND
       NEW.profile_values IS NOT DISTINCT FROM OLD.profile_values AND
       NEW.profile_sha256 IS NOT DISTINCT FROM OLD.profile_sha256 THEN
        RETURN NEW;
    END IF;
    SELECT plan.target_schema_id INTO v_target_schema_id
    FROM iam_profile_backfill_plans AS plan
    WHERE plan.source_schema_id=OLD.profile_schema_version_id
      AND plan.resource_kind=TG_ARGV[0]
      AND plan.state IN ('planned','running','verified');
    IF FOUND AND NEW.profile_schema_version_id IS DISTINCT FROM v_target_schema_id THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='IAM profile source mutation is fenced for backfill';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_guard_profile_backfill_source_write_v1() FROM PUBLIC;

CREATE TRIGGER iam_users_profile_backfill_insert_guard
BEFORE INSERT ON iam_users FOR EACH ROW
EXECUTE FUNCTION iam_guard_profile_backfill_source_write_v1('user');
ALTER TABLE iam_users ENABLE ALWAYS TRIGGER iam_users_profile_backfill_insert_guard;

CREATE TRIGGER iam_users_profile_backfill_update_guard
BEFORE UPDATE OF profile_schema_version_id,profile_values,profile_sha256 ON iam_users
FOR EACH ROW EXECUTE FUNCTION iam_guard_profile_backfill_source_write_v1('user');
ALTER TABLE iam_users ENABLE ALWAYS TRIGGER iam_users_profile_backfill_update_guard;

CREATE TRIGGER iam_organizations_profile_backfill_insert_guard
BEFORE INSERT ON iam_organizations FOR EACH ROW
EXECUTE FUNCTION iam_guard_profile_backfill_source_write_v1('organization');
ALTER TABLE iam_organizations ENABLE ALWAYS TRIGGER iam_organizations_profile_backfill_insert_guard;

CREATE TRIGGER iam_organizations_profile_backfill_update_guard
BEFORE UPDATE OF profile_schema_version_id,profile_values,profile_sha256 ON iam_organizations
FOR EACH ROW EXECUTE FUNCTION iam_guard_profile_backfill_source_write_v1('organization');
ALTER TABLE iam_organizations ENABLE ALWAYS TRIGGER iam_organizations_profile_backfill_update_guard;

CREATE TRIGGER iam_roles_profile_backfill_insert_guard
BEFORE INSERT ON iam_roles FOR EACH ROW
EXECUTE FUNCTION iam_guard_profile_backfill_source_write_v1('role');
ALTER TABLE iam_roles ENABLE ALWAYS TRIGGER iam_roles_profile_backfill_insert_guard;

CREATE TRIGGER iam_roles_profile_backfill_update_guard
BEFORE UPDATE OF profile_schema_version_id,profile_values,profile_sha256 ON iam_roles
FOR EACH ROW EXECUTE FUNCTION iam_guard_profile_backfill_source_write_v1('role');
ALTER TABLE iam_roles ENABLE ALWAYS TRIGGER iam_roles_profile_backfill_update_guard;

