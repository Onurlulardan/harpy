-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_backfill_forms_completion_v1(
    p_contract_version text
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_schema_job platform.schema_jobs%ROWTYPE;
    v_completion_job_id uuid;
BEGIN
    IF p_contract_version IS DISTINCT FROM
       'forms-schema-completion-backfill.v1' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'schema completion backfill contract is invalid';
    END IF;
    IF EXISTS (
        SELECT 1
        FROM platform.schema_jobs AS candidate
        WHERE candidate.owner = 'forms'
          AND candidate.resource_kind = 'form_storage'
          AND candidate.plan_contract_version = 'forms-schema-plan.v2'
          AND candidate.compiler_version = '2.0.0'
          AND candidate.state IN ('succeeded', 'failed', 'cancelled')
          AND (
              candidate.submission_authority_id IS NULL
              OR candidate.submission_authority_group_id IS NULL
              OR candidate.submission_group_generation IS NULL
              OR candidate.submission_billing_generation IS NULL
              OR candidate.submission_forms_generation IS NULL
          )
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'terminal V2 schema job lacks frozen authority';
    END IF;

    FOR v_schema_job IN
        SELECT candidate.*
        FROM platform.schema_jobs AS candidate
        WHERE candidate.owner = 'forms'
          AND candidate.resource_kind = 'form_storage'
          AND candidate.plan_contract_version = 'forms-schema-plan.v2'
          AND candidate.compiler_version = '2.0.0'
          AND candidate.state IN ('succeeded', 'failed', 'cancelled')
        ORDER BY candidate.job_id
        FOR SHARE
    LOOP
        v_completion_job_id :=
            platform.platform_async_enqueue_forms_schema_completion_v1(
                v_schema_job.job_id,
                v_schema_job.job_id,
                v_schema_job.submission_authority_id,
                v_schema_job.submission_authority_group_id,
                v_schema_job.submission_group_generation,
                v_schema_job.submission_billing_generation,
                v_schema_job.submission_forms_generation
            );
        IF pg_catalog.uuid_extract_version(v_completion_job_id) IS DISTINCT FROM 7 THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'schema completion backfill job is invalid';
        END IF;
    END LOOP;
    RETURN;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
platform.platform_migrator_backfill_forms_completion_v1(text)
FROM PUBLIC;

SELECT platform.platform_migrator_backfill_forms_completion_v1(
    'forms-schema-completion-backfill.v1'
);
