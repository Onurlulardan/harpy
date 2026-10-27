-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_enqueue_forms_schema_runtime_v2(
    p_schema_job_id uuid,
    p_schema_job_generation bigint
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_schema_job platform.schema_jobs%ROWTYPE;
    v_step platform.schema_job_steps%ROWTYPE;
    v_payload_ref text;
    v_payload_sha256 text;
    v_dedupe_key text;
    v_now timestamptz;
    v_run_at timestamptz;
    v_job_id uuid;
BEGIN
    IF SESSION_USER NOT IN ('harpy_migrator','harpy_v2_runtime')
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.uuid_extract_version(p_schema_job_id) IS DISTINCT FROM 7
       OR p_schema_job_generation < 1 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Forms schema runtime enqueue authority is invalid';
    END IF;
    SELECT candidate.* INTO STRICT v_schema_job
    FROM platform.schema_jobs AS candidate
    JOIN platform.schema_job_steps AS step
      ON step.job_id=candidate.job_id
     AND step.ordinal=candidate.next_step_ordinal
    WHERE candidate.job_id=p_schema_job_id
      AND candidate.generation=p_schema_job_generation
      AND candidate.owner='forms'
      AND candidate.resource_kind='form_storage'
      AND candidate.plan_contract_version='forms-schema-plan.v2'
      AND candidate.compiler_version='2.0.0'
      AND candidate.state='awaiting_runtime'
      AND candidate.next_executor='forms_runtime'
      AND step.executor_kind='forms_runtime'
      AND step.state IN ('pending','retryable')
    FOR SHARE OF candidate,step;
    SELECT step.* INTO STRICT v_step
    FROM platform.schema_job_steps AS step
    WHERE step.job_id=v_schema_job.job_id
      AND step.ordinal=v_schema_job.next_step_ordinal;
    IF v_schema_job.submission_authority_id IS NULL
       OR v_schema_job.submission_authority_group_id IS NULL
       OR v_schema_job.submission_group_generation IS NULL
       OR v_schema_job.submission_forms_generation IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'Forms schema runtime frozen authority is missing';
    END IF;
    v_payload_ref := 'forms-schema-runtime/'||p_schema_job_id::text||'/'||
        p_schema_job_generation::text||'/'||v_step.ordinal::text;
    v_payload_sha256 := pg_catalog.encode(pg_catalog.sha256(
        pg_catalog.convert_to(v_payload_ref,'UTF8')),'hex');
    v_dedupe_key := 'forms.schema-evolution.runtime:'||p_schema_job_id::text||
        ':'||p_schema_job_generation::text;
    v_now := pg_catalog.clock_timestamp();
    v_run_at := GREATEST(v_schema_job.available_at,v_now);
    v_job_id := pg_catalog.uuidv7();
    INSERT INTO platform.jobs (
        id,tenant_id,scope_kind,scope_id,authority_group_id,
        authority_group_generation,authority_domain,cutover_generation,
        writer_origin,kind,payload_revision,payload_ref,payload_sha256,
        dedupe_key,effect_class,priority,run_at,state,attempts,max_attempts,
        lease_generation,created_at,updated_at
    ) VALUES (
        v_job_id,v_schema_job.tenant_id,'tenant',v_schema_job.tenant_id,
        v_schema_job.submission_authority_group_id,
        v_schema_job.submission_group_generation,'forms',
        v_schema_job.submission_forms_generation,'v2',
        'forms.schema-evolution.runtime',1,v_payload_ref,v_payload_sha256,
        v_dedupe_key,'database_transactional',100,v_run_at,'queued',0,8,0,
        v_now,v_now
    ) ON CONFLICT ON CONSTRAINT jobs_dedupe_unique DO NOTHING
    RETURNING id INTO v_job_id;
    IF NOT FOUND THEN
        SELECT existing.id INTO STRICT v_job_id FROM platform.jobs AS existing
        WHERE existing.tenant_id=v_schema_job.tenant_id
          AND existing.scope_kind='tenant'
          AND existing.scope_id=v_schema_job.tenant_id
          AND existing.authority_group_id=v_schema_job.submission_authority_group_id
          AND existing.authority_group_generation=
              v_schema_job.submission_group_generation
          AND existing.authority_domain='forms'
          AND existing.cutover_generation=v_schema_job.submission_forms_generation
          AND existing.writer_origin='v2'
          AND existing.kind='forms.schema-evolution.runtime'
          AND existing.payload_revision=1
          AND existing.payload_ref=v_payload_ref
          AND existing.payload_sha256=v_payload_sha256
          AND existing.dedupe_key=v_dedupe_key
          AND existing.effect_class='database_transactional'
          AND existing.effect_idempotency_key IS NULL
          AND existing.max_attempts=8
        FOR SHARE;
    END IF;
    RETURN v_job_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
platform.platform_async_enqueue_forms_schema_runtime_v2(uuid,bigint)
FROM PUBLIC;
