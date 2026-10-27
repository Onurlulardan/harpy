-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_finish_forms_runtime_v2(
    p_schema_job_id uuid, p_runtime_job_id uuid, p_runtime_attempt_id uuid,
    p_job_generation bigint, p_step_ordinal integer,
    p_step_generation bigint, p_ready boolean, p_retry_after timestamptz
)
RETURNS TABLE (job_state text, job_generation bigint)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.schema_jobs%ROWTYPE;
    v_step platform.schema_job_steps%ROWTYPE;
    v_next platform.schema_job_steps%ROWTYPE;
    v_now timestamptz;
    v_step_state text;
    v_error_code text;
    v_outcome_sha256 text;
    v_next_state text;
    v_next_executor text;
    v_job_generation bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.uuid_extract_version(p_schema_job_id) IS DISTINCT FROM 7
       OR pg_catalog.uuid_extract_version(p_runtime_job_id) IS DISTINCT FROM 7
       OR pg_catalog.uuid_extract_version(p_runtime_attempt_id) IS DISTINCT FROM 7
       OR p_job_generation < 1 OR p_step_ordinal < 0 OR p_step_generation < 1
       OR p_ready IS NULL OR (p_ready AND p_retry_after IS NOT NULL)
       OR (NOT p_ready AND (p_retry_after IS NULL
           OR NOT pg_catalog.isfinite(p_retry_after))) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Forms schema runtime finish authority is invalid';
    END IF;
    SELECT candidate.* INTO v_job
    FROM platform.schema_jobs AS candidate
    JOIN platform.schema_job_steps AS step
      ON step.job_id=candidate.job_id
     AND step.ordinal=candidate.next_step_ordinal
    WHERE candidate.job_id=p_schema_job_id
      AND candidate.state='running'
      AND candidate.next_executor='forms_runtime'
      AND candidate.generation=p_job_generation
      AND candidate.lease_id=p_runtime_job_id
      AND candidate.lease_owner_id=p_runtime_attempt_id
      AND candidate.next_step_ordinal=p_step_ordinal
      AND step.executor_kind='forms_runtime' AND step.state='running'
      AND step.generation=p_step_generation
    FOR UPDATE OF candidate,step;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='Forms schema runtime finish fence changed';
    END IF;
    SELECT step.* INTO STRICT v_step
    FROM platform.schema_job_steps AS step
    WHERE step.job_id=p_schema_job_id AND step.ordinal=p_step_ordinal;
    v_now := pg_catalog.clock_timestamp();
    IF v_job.lease_expires_at <= v_now OR
       (NOT p_ready AND (p_retry_after <= v_now
        OR p_retry_after > v_now+interval '31 days')) THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='Forms schema runtime finish time fence changed';
    END IF;
    v_step_state := CASE WHEN p_ready THEN 'succeeded' ELSE 'retryable' END;
    v_error_code := CASE WHEN p_ready THEN NULL ELSE 'retention_gate_waiting' END;
    v_outcome_sha256 := pg_catalog.encode(pg_catalog.sha256(
        pg_catalog.convert_to('harpy:v2:forms-schema-runtime:v1|'||
            p_schema_job_id::text||'|'||p_step_ordinal::text||'|'||
            v_step.attempt::text||'|'||v_step_state,'UTF8')),'hex');
    INSERT INTO platform.schema_job_step_attempts (
        job_id,step_ordinal,attempt,plan_sha256,step_sha256,lease_id,
        lease_owner_id,executor_kind,state,started_at,finished_at,
        outcome_sha256,error_code,redacted_detail_sha256,recorded_at
    ) VALUES (p_schema_job_id,p_step_ordinal,v_step.attempt,v_job.plan_sha256,
        v_step.step_sha256,p_runtime_job_id,p_runtime_attempt_id,
        'forms_runtime',v_step_state,v_step.started_at,v_now,v_outcome_sha256,
        v_error_code,CASE WHEN p_ready THEN NULL ELSE v_outcome_sha256 END,v_now);
    UPDATE platform.schema_job_steps AS step SET state=v_step_state,
        generation=step.generation+1,
        next_executor=CASE WHEN p_ready THEN 'none' ELSE 'forms_runtime' END,
        finished_at=v_now,error_code=v_error_code,updated_at=v_now
    WHERE step.job_id=p_schema_job_id AND step.ordinal=p_step_ordinal
      AND step.generation=p_step_generation AND step.state='running';
    IF NOT p_ready THEN
        UPDATE platform.schema_jobs AS candidate SET state='awaiting_runtime',
            generation=candidate.generation+1,next_executor='forms_runtime',
            available_at=p_retry_after,lease_id=NULL,lease_owner_id=NULL,
            leased_at=NULL,lease_heartbeat_at=NULL,lease_expires_at=NULL,
            updated_at=v_now
        WHERE candidate.job_id=p_schema_job_id
          AND candidate.generation=p_job_generation
        RETURNING candidate.state,candidate.generation
        INTO v_next_state,v_job_generation;
        RETURN QUERY SELECT v_next_state,v_job_generation;
        RETURN;
    END IF;
    IF p_step_ordinal+1=v_job.step_count THEN
        UPDATE platform.schema_jobs AS candidate SET state='succeeded',
            generation=candidate.generation+1,next_step_ordinal=step_count,
            next_executor='none',lease_id=NULL,lease_owner_id=NULL,
            leased_at=NULL,lease_heartbeat_at=NULL,lease_expires_at=NULL,
            updated_at=v_now,finished_at=v_now
        WHERE candidate.job_id=p_schema_job_id
          AND candidate.generation=p_job_generation
        RETURNING candidate.state,candidate.generation
        INTO v_next_state,v_job_generation;
    ELSE
        SELECT step.* INTO STRICT v_next
        FROM platform.schema_job_steps AS step
        WHERE step.job_id=p_schema_job_id AND step.ordinal=p_step_ordinal+1
          AND step.state='pending' FOR SHARE;
        v_next_executor := v_next.executor_kind;
        v_next_state := CASE WHEN v_next_executor='forms_runtime'
            THEN 'awaiting_runtime' ELSE 'queued' END;
        UPDATE platform.schema_jobs AS candidate SET state=v_next_state,
            generation=candidate.generation+1,
            next_step_ordinal=candidate.next_step_ordinal+1,
            next_executor=v_next_executor,available_at=v_now,
            lease_id=NULL,lease_owner_id=NULL,leased_at=NULL,
            lease_heartbeat_at=NULL,lease_expires_at=NULL,updated_at=v_now
        WHERE candidate.job_id=p_schema_job_id
          AND candidate.generation=p_job_generation
        RETURNING candidate.generation INTO v_job_generation;
    END IF;
    RETURN QUERY SELECT v_next_state,v_job_generation;
END;
$function$;

REVOKE ALL ON FUNCTION platform.platform_migrator_finish_forms_runtime_v2(
    uuid,uuid,uuid,bigint,integer,bigint,boolean,timestamptz
) FROM PUBLIC;

CREATE FUNCTION platform.platform_migrator_backfill_forms_runtime_v2(
    p_contract_version text
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.schema_jobs%ROWTYPE;
    v_runtime_job_id uuid;
BEGIN
    IF p_contract_version IS DISTINCT FROM
       'forms-schema-runtime-backfill.v2' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='schema runtime backfill contract is invalid';
    END IF;
    FOR v_job IN
        SELECT candidate.* FROM platform.schema_jobs AS candidate
        WHERE candidate.owner='forms'
          AND candidate.resource_kind='form_storage'
          AND candidate.plan_contract_version='forms-schema-plan.v2'
          AND candidate.compiler_version='2.0.0'
          AND candidate.state='awaiting_runtime'
          AND candidate.next_executor='forms_runtime'
        ORDER BY candidate.job_id FOR SHARE
    LOOP
        v_runtime_job_id :=
            platform.platform_async_enqueue_forms_schema_runtime_v2(
                v_job.job_id,v_job.generation);
        IF pg_catalog.uuid_extract_version(v_runtime_job_id) IS DISTINCT FROM 7 THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='schema runtime backfill job is invalid';
        END IF;
    END LOOP;
END;
$function$;

REVOKE ALL ON FUNCTION
platform.platform_migrator_backfill_forms_runtime_v2(text)
FROM PUBLIC;

SELECT platform.platform_migrator_backfill_forms_runtime_v2(
    'forms-schema-runtime-backfill.v2');

REVOKE EXECUTE ON FUNCTION
platform.platform_migrator_claim_forms_runtime_v1(
    uuid,bigint,integer,uuid,uuid
) FROM harpy_runtime;

REVOKE EXECUTE ON FUNCTION
platform.platform_migrator_finish_forms_runtime_v1(
    uuid,uuid,uuid,bigint,integer,bigint,boolean,timestamptz
) FROM harpy_runtime;

GRANT EXECUTE ON FUNCTION
platform.platform_migrator_claim_forms_runtime_v2(
    uuid,bigint,integer,uuid,uuid
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION
platform.platform_migrator_finish_forms_runtime_v2(
    uuid,uuid,uuid,bigint,integer,bigint,boolean,timestamptz
) TO harpy_runtime;
-- +goose StatementEnd
