-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_claim_forms_runtime_v2(
    p_schema_job_id uuid, p_expected_job_generation bigint,
    p_expected_step_ordinal integer, p_runtime_job_id uuid,
    p_runtime_attempt_id uuid
)
RETURNS TABLE (
    schema_job_id uuid, tenant_id uuid, form_id uuid,
    resource_revision bigint, job_generation bigint, step_ordinal integer,
    step_generation bigint, step_attempt integer, operation_id text,
    identifier_mapping jsonb, arguments jsonb, plan_sha256 text,
    step_sha256 text
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.schema_jobs%ROWTYPE;
    v_step platform.schema_job_steps%ROWTYPE;
    v_now timestamptz;
    v_job_generation bigint;
    v_step_generation bigint;
    v_step_attempt integer;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.uuid_extract_version(p_schema_job_id) IS DISTINCT FROM 7
       OR p_expected_job_generation < 1 OR p_expected_step_ordinal < 0
       OR pg_catalog.uuid_extract_version(p_runtime_job_id) IS DISTINCT FROM 7
       OR pg_catalog.uuid_extract_version(p_runtime_attempt_id) IS DISTINCT FROM 7 THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Forms schema runtime claim authority is invalid';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM platform.jobs AS runtime_job
        JOIN platform.job_attempts AS attempt
          ON attempt.job_id=runtime_job.id
         AND attempt.attempt_id=p_runtime_attempt_id
         AND attempt.attempt=runtime_job.attempts
         AND attempt.lease_generation=runtime_job.lease_generation
         AND attempt.lease_owner_id=runtime_job.lease_owner_id
        WHERE runtime_job.id=p_runtime_job_id
          AND runtime_job.state='running'
          AND runtime_job.kind='forms.schema-evolution.runtime'
          AND runtime_job.payload_revision=1
          AND runtime_job.effect_class='database_transactional'
          AND runtime_job.payload_ref='forms-schema-runtime/'||
              p_schema_job_id::text||'/'||p_expected_job_generation::text||
              '/'||p_expected_step_ordinal::text
    ) THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='Forms schema runtime async lease changed';
    END IF;
    SELECT candidate.* INTO v_job
    FROM platform.schema_jobs AS candidate
    JOIN platform.schema_job_steps AS step
      ON step.job_id=candidate.job_id
     AND step.ordinal=candidate.next_step_ordinal
    WHERE candidate.job_id=p_schema_job_id
      AND candidate.generation=p_expected_job_generation
      AND candidate.next_step_ordinal=p_expected_step_ordinal
      AND candidate.owner='forms' AND candidate.resource_kind='form_storage'
      AND candidate.plan_contract_version='forms-schema-plan.v2'
      AND candidate.compiler_version='2.0.0'
      AND candidate.state='awaiting_runtime'
      AND candidate.next_executor='forms_runtime'
      AND step.executor_kind='forms_runtime'
      AND step.state IN ('pending','retryable')
    FOR UPDATE OF candidate,step;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='Forms schema runtime platform fence changed';
    END IF;
    SELECT step.* INTO STRICT v_step
    FROM platform.schema_job_steps AS step
    WHERE step.job_id=p_schema_job_id
      AND step.ordinal=p_expected_step_ordinal;
    v_now := pg_catalog.clock_timestamp();
    UPDATE platform.schema_jobs AS candidate SET state='leased',
        generation=candidate.generation+1,attempt=candidate.attempt+1,
        lease_id=p_runtime_job_id,lease_owner_id=p_runtime_attempt_id,
        leased_at=v_now,lease_heartbeat_at=v_now,
        lease_expires_at=v_now+interval '60 seconds',updated_at=v_now
    WHERE candidate.job_id=p_schema_job_id
      AND candidate.generation=p_expected_job_generation
    RETURNING candidate.generation INTO v_job_generation;
    UPDATE platform.schema_job_steps AS step SET state='running',
        generation=step.generation+1,attempt=step.attempt+1,
        started_at=v_now,finished_at=NULL,error_code=NULL,updated_at=v_now
    WHERE step.job_id=p_schema_job_id AND step.ordinal=p_expected_step_ordinal
      AND step.generation=v_step.generation
      AND step.state IN ('pending','retryable')
    RETURNING step.generation,step.attempt
    INTO v_step_generation,v_step_attempt;
    UPDATE platform.schema_jobs AS candidate SET state='running',
        generation=candidate.generation+1,updated_at=v_now
    WHERE candidate.job_id=p_schema_job_id
      AND candidate.generation=v_job_generation
      AND candidate.state='leased'
    RETURNING candidate.generation INTO v_job_generation;
    RETURN QUERY SELECT v_job.job_id,v_job.tenant_id,v_job.resource_id,
        v_job.resource_revision,v_job_generation,v_step.ordinal,
        v_step_generation,v_step_attempt,v_step.operation_id,
        v_step.identifier_mapping,v_step.arguments,v_job.plan_sha256,
        v_step.step_sha256;
END;
$function$;

REVOKE ALL ON FUNCTION platform.platform_migrator_claim_forms_runtime_v2(
    uuid,bigint,integer,uuid,uuid
) FROM PUBLIC;

CREATE FUNCTION platform.platform_migrator_enqueue_forms_runtime_v2()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_runtime_job_id uuid;
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE'
       OR TG_TABLE_SCHEMA IS DISTINCT FROM 'platform'
       OR TG_TABLE_NAME IS DISTINCT FROM 'schema_jobs' THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='schema runtime trigger target is invalid';
    END IF;
    IF NEW.owner='forms' AND NEW.resource_kind='form_storage'
       AND NEW.plan_contract_version='forms-schema-plan.v2'
       AND NEW.compiler_version='2.0.0'
       AND NEW.state='awaiting_runtime'
       AND NEW.next_executor='forms_runtime'
       AND (OLD.state IS DISTINCT FROM NEW.state
            OR OLD.generation IS DISTINCT FROM NEW.generation) THEN
        v_runtime_job_id :=
            platform.platform_async_enqueue_forms_schema_runtime_v2(
                NEW.job_id,NEW.generation);
        IF pg_catalog.uuid_extract_version(v_runtime_job_id) IS DISTINCT FROM 7 THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='schema runtime durable job is invalid';
        END IF;
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION
platform.platform_migrator_enqueue_forms_runtime_v2()
FROM PUBLIC;

ALTER TABLE platform.schema_jobs
    DISABLE TRIGGER schema_jobs_forms_runtime_v1;

CREATE TRIGGER schema_jobs_forms_runtime_v2
AFTER UPDATE ON platform.schema_jobs
FOR EACH ROW
EXECUTE FUNCTION platform.platform_migrator_enqueue_forms_runtime_v2();
ALTER TABLE platform.schema_jobs
    ENABLE ALWAYS TRIGGER schema_jobs_forms_runtime_v2;
-- +goose StatementEnd
