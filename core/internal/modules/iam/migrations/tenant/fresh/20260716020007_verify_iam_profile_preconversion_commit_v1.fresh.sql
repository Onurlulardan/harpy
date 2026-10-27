-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_verify_profile_preconversion_commit_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_target_schema_id uuid;
    v_current_schema_id uuid;
    v_current_revision bigint;
    v_current_sha256 bytea;
BEGIN
    SELECT plan.target_schema_id INTO v_target_schema_id
    FROM iam_profile_backfill_plans AS plan
    JOIN iam_profile_backfill_pages AS page
      ON page.plan_id = plan.id AND page.page_number = NEW.page_number
    JOIN iam_profile_schema_versions AS actual
      ON actual.id = NEW.actual_source_schema_id
     AND actual.resource_kind = NEW.resource_kind
     AND actual.state IN ('active','retired')
    JOIN iam_profile_schema_versions AS source
      ON source.id = plan.source_schema_id
     AND source.resource_kind = plan.resource_kind
     AND actual.version <= source.version
    JOIN iam_profile_schema_versions AS target
      ON target.id = plan.target_schema_id
     AND target.resource_kind = plan.resource_kind
     AND target.state = 'draft'
    WHERE plan.id = NEW.plan_id
      AND plan.resource_kind = NEW.resource_kind
      AND plan.state IN ('planned','running')
      AND (page.after_resource_id IS NULL OR
           NEW.resource_id > page.after_resource_id)
      AND NEW.resource_id <= page.through_resource_id;
    IF NEW.resource_kind = 'user' THEN
        SELECT resource.profile_schema_version_id, resource.revision,
               resource.profile_sha256
        INTO v_current_schema_id, v_current_revision, v_current_sha256
        FROM iam_users AS resource WHERE resource.id = NEW.resource_id;
    ELSIF NEW.resource_kind = 'organization' THEN
        SELECT resource.profile_schema_version_id, resource.revision,
               resource.profile_sha256
        INTO v_current_schema_id, v_current_revision, v_current_sha256
        FROM iam_organizations AS resource WHERE resource.id = NEW.resource_id;
    ELSIF NEW.resource_kind = 'role' THEN
        SELECT resource.profile_schema_version_id, resource.revision,
               resource.profile_sha256
        INTO v_current_schema_id, v_current_revision, v_current_sha256
        FROM iam_roles AS resource WHERE resource.id = NEW.resource_id;
    END IF;
    IF v_target_schema_id IS NULL OR
       v_current_schema_id IS DISTINCT FROM v_target_schema_id OR
       v_current_revision < NEW.first_target_revision OR
       (v_current_revision = NEW.first_target_revision AND
        v_current_sha256 IS DISTINCT FROM NEW.first_target_sha256) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM profile preconversion did not commit its target write';
    END IF;
    RETURN NULL;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_verify_profile_preconversion_commit_v1()
    FROM PUBLIC;

CREATE CONSTRAINT TRIGGER iam_profile_preconversions_commit_guard
AFTER INSERT ON iam_profile_backfill_preconversions
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION iam_verify_profile_preconversion_commit_v1();
ALTER TABLE iam_profile_backfill_preconversions ENABLE ALWAYS TRIGGER
    iam_profile_preconversions_commit_guard;
