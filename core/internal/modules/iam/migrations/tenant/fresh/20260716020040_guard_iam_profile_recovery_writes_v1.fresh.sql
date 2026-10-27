-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_guard_profile_backfill_source_write_v4()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_plan_id uuid;
    v_target_schema_id uuid;
    v_source_version bigint;
    v_old_version bigint;
    v_worker_capability boolean;
BEGIN
    IF TG_OP = 'UPDATE' AND
       NEW.profile_schema_version_id IS NOT DISTINCT FROM
           OLD.profile_schema_version_id AND
       NEW.profile_values IS NOT DISTINCT FROM OLD.profile_values AND
       NEW.profile_sha256 IS NOT DISTINCT FROM OLD.profile_sha256 THEN
        RETURN NEW;
    END IF;
    SELECT plan.id, plan.target_schema_id, source.version
    INTO v_plan_id, v_target_schema_id, v_source_version
    FROM iam_profile_backfill_plans AS plan
    JOIN iam_profile_schema_versions AS source
      ON source.id = plan.source_schema_id
     AND source.resource_kind = plan.resource_kind AND source.state = 'active'
    JOIN iam_profile_schema_versions AS target
      ON target.id = plan.target_schema_id
     AND target.resource_kind = plan.resource_kind AND target.state = 'draft'
    WHERE plan.resource_kind = TG_ARGV[0]
      AND (plan.state IN ('planned','running') OR
           (plan.recovery_of_plan_id IS NULL AND plan.state = 'verified'))
    FOR SHARE OF plan, source, target;
    IF NOT FOUND THEN
        IF EXISTS (
            SELECT 1 FROM iam_profile_backfill_plans AS fenced
            JOIN iam_profile_schema_versions AS source
              ON source.id = fenced.source_schema_id
             AND source.resource_kind = fenced.resource_kind
             AND source.state = 'active'
            JOIN iam_profile_schema_versions AS target
              ON target.id = fenced.target_schema_id
             AND target.resource_kind = fenced.resource_kind
             AND target.state = 'draft'
            WHERE fenced.resource_kind = TG_ARGV[0]
              AND fenced.state = 'fenced'
        ) THEN
            RAISE EXCEPTION USING ERRCODE = 'P5741',
                MESSAGE = 'IAM profile domain is fenced pending recovery';
        END IF;
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
    IF v_old_version IS NULL OR v_old_version > v_source_version OR
       NEW.revision IS DISTINCT FROM OLD.revision + 1 THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile lineage transition is fenced';
    END IF;
    v_worker_capability := iam_profile_worker_capability_consume_v1(
        v_plan_id, TG_ARGV[0], NEW.id, OLD.profile_schema_version_id,
        OLD.revision, OLD.profile_sha256, NEW.profile_schema_version_id,
        NEW.profile_sha256
    );
    IF v_worker_capability THEN
        RETURN NEW;
    END IF;
    IF EXISTS (
        SELECT 1 FROM iam_profile_backfill_preconversions AS receipt
        WHERE receipt.plan_id = v_plan_id
          AND receipt.resource_kind = TG_ARGV[0]
          AND receipt.resource_id = NEW.id
          AND receipt.actual_source_schema_id = OLD.profile_schema_version_id
          AND receipt.source_revision = OLD.revision
          AND receipt.source_sha256 = OLD.profile_sha256
          AND receipt.first_target_revision = NEW.revision
          AND receipt.first_target_sha256 = NEW.profile_sha256
    ) THEN
        RETURN NEW;
    END IF;
    RAISE EXCEPTION USING ERRCODE = '40001',
        MESSAGE = 'IAM profile source transition lacks authority evidence';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_guard_profile_backfill_source_write_v4()
    FROM PUBLIC;

CREATE OR REPLACE TRIGGER iam_users_profile_backfill_insert_guard
BEFORE INSERT ON iam_users FOR EACH ROW
EXECUTE FUNCTION iam_guard_profile_backfill_source_write_v4('user');
CREATE OR REPLACE TRIGGER iam_users_profile_backfill_update_guard
BEFORE UPDATE OF profile_schema_version_id,profile_values,profile_sha256
ON iam_users FOR EACH ROW
EXECUTE FUNCTION iam_guard_profile_backfill_source_write_v4('user');

CREATE OR REPLACE TRIGGER iam_organizations_profile_backfill_insert_guard
BEFORE INSERT ON iam_organizations FOR EACH ROW
EXECUTE FUNCTION iam_guard_profile_backfill_source_write_v4('organization');
CREATE OR REPLACE TRIGGER iam_organizations_profile_backfill_update_guard
BEFORE UPDATE OF profile_schema_version_id,profile_values,profile_sha256
ON iam_organizations FOR EACH ROW
EXECUTE FUNCTION iam_guard_profile_backfill_source_write_v4('organization');

CREATE OR REPLACE TRIGGER iam_roles_profile_backfill_insert_guard
BEFORE INSERT ON iam_roles FOR EACH ROW
EXECUTE FUNCTION iam_guard_profile_backfill_source_write_v4('role');
CREATE OR REPLACE TRIGGER iam_roles_profile_backfill_update_guard
BEFORE UPDATE OF profile_schema_version_id,profile_values,profile_sha256
ON iam_roles FOR EACH ROW
EXECUTE FUNCTION iam_guard_profile_backfill_source_write_v4('role');

ALTER TABLE iam_users ENABLE ALWAYS TRIGGER
    iam_users_profile_backfill_insert_guard;
ALTER TABLE iam_users ENABLE ALWAYS TRIGGER
    iam_users_profile_backfill_update_guard;
ALTER TABLE iam_organizations ENABLE ALWAYS TRIGGER
    iam_organizations_profile_backfill_insert_guard;
ALTER TABLE iam_organizations ENABLE ALWAYS TRIGGER
    iam_organizations_profile_backfill_update_guard;
ALTER TABLE iam_roles ENABLE ALWAYS TRIGGER
    iam_roles_profile_backfill_insert_guard;
ALTER TABLE iam_roles ENABLE ALWAYS TRIGGER
    iam_roles_profile_backfill_update_guard;
