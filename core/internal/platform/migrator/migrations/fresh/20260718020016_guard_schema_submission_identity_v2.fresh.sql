-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_guard_schema_submission_identity_v2()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
BEGIN
    IF ROW(
        NEW.canonical_plan_bytes,NEW.submission_authority_id,
        NEW.submission_authority_group_id,NEW.submission_group_generation,
        NEW.submission_billing_generation,NEW.submission_forms_generation
    ) IS DISTINCT FROM ROW(
        OLD.canonical_plan_bytes,OLD.submission_authority_id,
        OLD.submission_authority_group_id,OLD.submission_group_generation,
        OLD.submission_billing_generation,OLD.submission_forms_generation
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='schema submission authority identity is immutable';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
platform.platform_migrator_guard_schema_submission_identity_v2()
FROM PUBLIC;

CREATE TRIGGER schema_jobs_guard_submission_identity_v2
BEFORE UPDATE ON platform.schema_jobs
FOR EACH ROW
EXECUTE FUNCTION platform.platform_migrator_guard_schema_submission_identity_v2();

ALTER TABLE platform.schema_jobs
    ENABLE ALWAYS TRIGGER schema_jobs_guard_submission_identity_v2;
