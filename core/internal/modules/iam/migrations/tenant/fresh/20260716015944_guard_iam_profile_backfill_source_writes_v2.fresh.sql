-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_guard_profile_backfill_source_write_v2()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_target_schema_id uuid;
    v_source_version bigint;
    v_old_version bigint;
BEGIN
    SELECT plan.target_schema_id, source.version
    INTO v_target_schema_id, v_source_version
    FROM iam_profile_backfill_plans AS plan
    JOIN iam_profile_schema_versions AS source
      ON source.id = plan.source_schema_id
     AND source.resource_kind = plan.resource_kind AND source.state = 'active'
    JOIN iam_profile_schema_versions AS target
      ON target.id = plan.target_schema_id
     AND target.resource_kind = plan.resource_kind AND target.state = 'draft'
    WHERE plan.resource_kind = TG_ARGV[0]
      AND plan.state IN ('planned','running','verified');
    IF NOT FOUND THEN
        RETURN NEW;
    END IF;
    IF TG_OP = 'UPDATE' AND
       NEW.profile_schema_version_id IS NOT DISTINCT FROM OLD.profile_schema_version_id AND
       NEW.profile_values IS NOT DISTINCT FROM OLD.profile_values AND
       NEW.profile_sha256 IS NOT DISTINCT FROM OLD.profile_sha256 THEN
        RETURN NEW;
    END IF;
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       NEW.profile_schema_version_id IS DISTINCT FROM v_target_schema_id THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile resource kind is fenced for backfill';
    END IF;
    IF TG_OP = 'INSERT' OR
       OLD.profile_schema_version_id IS NOT DISTINCT FROM v_target_schema_id THEN
        RETURN NEW;
    END IF;
    SELECT actual.version INTO v_old_version
    FROM iam_profile_schema_versions AS actual
    WHERE actual.id = OLD.profile_schema_version_id
      AND actual.resource_kind = TG_ARGV[0]
      AND actual.state IN ('active','retired');
    IF v_old_version IS NULL OR v_old_version > v_source_version THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile lineage transition is fenced';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_guard_profile_backfill_source_write_v2() FROM PUBLIC;

CREATE OR REPLACE TRIGGER iam_users_profile_backfill_insert_guard
BEFORE INSERT ON iam_users FOR EACH ROW
EXECUTE FUNCTION iam_guard_profile_backfill_source_write_v2('user');
ALTER TABLE iam_users ENABLE ALWAYS TRIGGER iam_users_profile_backfill_insert_guard;
CREATE OR REPLACE TRIGGER iam_users_profile_backfill_update_guard
BEFORE UPDATE OF profile_schema_version_id,profile_values,profile_sha256 ON iam_users
FOR EACH ROW EXECUTE FUNCTION iam_guard_profile_backfill_source_write_v2('user');
ALTER TABLE iam_users ENABLE ALWAYS TRIGGER iam_users_profile_backfill_update_guard;

CREATE OR REPLACE TRIGGER iam_organizations_profile_backfill_insert_guard
BEFORE INSERT ON iam_organizations FOR EACH ROW
EXECUTE FUNCTION iam_guard_profile_backfill_source_write_v2('organization');
ALTER TABLE iam_organizations ENABLE ALWAYS TRIGGER iam_organizations_profile_backfill_insert_guard;
CREATE OR REPLACE TRIGGER iam_organizations_profile_backfill_update_guard
BEFORE UPDATE OF profile_schema_version_id,profile_values,profile_sha256 ON iam_organizations
FOR EACH ROW EXECUTE FUNCTION iam_guard_profile_backfill_source_write_v2('organization');
ALTER TABLE iam_organizations ENABLE ALWAYS TRIGGER iam_organizations_profile_backfill_update_guard;

CREATE OR REPLACE TRIGGER iam_roles_profile_backfill_insert_guard
BEFORE INSERT ON iam_roles FOR EACH ROW
EXECUTE FUNCTION iam_guard_profile_backfill_source_write_v2('role');
ALTER TABLE iam_roles ENABLE ALWAYS TRIGGER iam_roles_profile_backfill_insert_guard;
CREATE OR REPLACE TRIGGER iam_roles_profile_backfill_update_guard
BEFORE UPDATE OF profile_schema_version_id,profile_values,profile_sha256 ON iam_roles
FOR EACH ROW EXECUTE FUNCTION iam_guard_profile_backfill_source_write_v2('role');
ALTER TABLE iam_roles ENABLE ALWAYS TRIGGER iam_roles_profile_backfill_update_guard;
