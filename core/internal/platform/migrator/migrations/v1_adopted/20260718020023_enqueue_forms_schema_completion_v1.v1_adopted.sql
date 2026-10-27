-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_enqueue_forms_completion_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_completion_job_id uuid;
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE'
       OR TG_TABLE_SCHEMA IS DISTINCT FROM 'platform'
       OR TG_TABLE_NAME IS DISTINCT FROM 'schema_jobs' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'schema completion trigger target is invalid';
    END IF;

    IF NEW.owner = 'forms'
       AND NEW.resource_kind = 'form_storage'
       AND NEW.plan_contract_version = 'forms-schema-plan.v2'
       AND NEW.compiler_version = '2.0.0'
       AND OLD.state NOT IN ('succeeded', 'failed', 'cancelled')
       AND NEW.state IN ('succeeded', 'failed', 'cancelled') THEN
        IF NEW.submission_authority_id IS NULL
           OR NEW.submission_authority_group_id IS NULL
           OR NEW.submission_group_generation IS NULL
           OR NEW.submission_billing_generation IS NULL
           OR NEW.submission_forms_generation IS NULL THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'schema completion frozen authority is missing';
        END IF;
        v_completion_job_id :=
            platform.platform_async_enqueue_forms_schema_completion_v1(
                NEW.job_id,
                NEW.job_id,
                NEW.submission_authority_id,
                NEW.submission_authority_group_id,
                NEW.submission_group_generation,
                NEW.submission_billing_generation,
                NEW.submission_forms_generation
            );
        IF pg_catalog.uuid_extract_version(v_completion_job_id) IS DISTINCT FROM 7 THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'schema completion durable job is invalid';
        END IF;
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
platform.platform_migrator_enqueue_forms_completion_v1()
FROM PUBLIC;

CREATE TRIGGER schema_jobs_forms_completion_v1
AFTER UPDATE ON platform.schema_jobs
FOR EACH ROW
EXECUTE FUNCTION platform.platform_migrator_enqueue_forms_completion_v1();

ALTER TABLE platform.schema_jobs
    ENABLE ALWAYS TRIGGER schema_jobs_forms_completion_v1;
