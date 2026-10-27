-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_dispose_stale_job_v1(
    p_job_id uuid,p_action text,p_root_id uuid,p_action_id uuid,
    p_result_job_id uuid,p_result_gap_root_id uuid,
    p_scope_kind text,p_scope_id uuid,p_group_id text,p_group_generation bigint,
    p_domain text,p_domain_generation bigint,p_result_payload_ref text,
    p_result_dedupe_key text,p_payload_canonical bytea,p_payload_sha256 text,
    p_reason text,p_ticket text,p_actor_id uuid,p_causation_id uuid,
    p_source_run_at text,p_result_run_at text,p_occurred_at text
)
RETURNS TABLE(completed_at timestamptz,action_id uuid,result_job_id uuid,
    result_root_id uuid,created boolean)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_result platform.jobs%ROWTYPE;
    v_now timestamptz;
    v_recorded_at timestamptz;
    v_replay_root_id uuid;
    v_replay_payload_ref text;
    v_replay_dedupe_key text;
    v_replay_run_at text;
    v_payload_ref text;
    v_code text;
    v_error_sha256 text;
    v_is_gap boolean;
BEGIN
    v_now := pg_catalog.transaction_timestamp();
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_job_id IS NULL OR p_root_id IS NULL OR p_action_id IS NULL OR
       p_action NOT IN ('adopt','cancel') OR
       NOT ((p_action = 'adopt' AND p_result_job_id IS NOT NULL) OR
            (p_action = 'cancel' AND p_result_job_id IS NULL AND
             p_result_gap_root_id IS NULL AND p_result_payload_ref IS NULL AND
             p_result_dedupe_key IS NULL AND p_result_run_at IS NULL)) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5687',
            MESSAGE = 'stale job disposition input is invalid';
    END IF;
    PERFORM platform.platform_async_assert_current_authority_v1(
        p_scope_kind,p_scope_id,p_group_id,p_group_generation,
        p_domain,p_domain_generation
    );
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id = p_job_id FOR UPDATE;
    IF NOT FOUND OR ROW(v_job.scope_kind,v_job.scope_id,
            v_job.authority_group_id,v_job.authority_domain)
        IS DISTINCT FROM ROW(p_scope_kind,p_scope_id,p_group_id,p_domain) OR
       p_group_generation < v_job.authority_group_generation OR
       p_domain_generation < v_job.cutover_generation OR
       (p_group_generation = v_job.authority_group_generation AND
        p_domain_generation = v_job.cutover_generation) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5688',
            MESSAGE = 'stale job disposition authority conflicts';
    END IF;
    IF p_source_run_at IS NULL OR p_source_run_at !~ 'Z$' OR
       p_source_run_at::timestamptz <> v_job.run_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P5687',
            MESSAGE = 'stale job source schedule is invalid';
    END IF;
    v_code := CASE p_action WHEN 'adopt' THEN
        'JOB_AUTHORITY_STALE_ADOPTED' ELSE 'JOB_AUTHORITY_STALE_CANCELLED' END;
    v_error_sha256 := pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        v_code || E'\n' || p_reason || E'\n' || p_ticket,'UTF8')),'hex');
    IF v_job.state = 'cancelled' THEN
        SELECT replay.recorded_at,replay.result_root_id,
            replay.result_payload_ref,replay.result_dedupe_key,
            replay.result_run_at
        INTO STRICT v_recorded_at,v_replay_root_id,v_replay_payload_ref,
            v_replay_dedupe_key,v_replay_run_at
        FROM platform.platform_async_verify_stale_job_disposition_replay_v1(
            v_job,p_root_id,p_action_id,p_result_job_id,p_group_generation,
            p_domain_generation,p_action,p_reason,p_ticket,p_actor_id,
            p_causation_id
        ) AS replay;
        RETURN QUERY SELECT v_recorded_at,p_action_id,p_result_job_id,
            v_replay_root_id,false;
        RETURN;
    END IF;
    IF v_job.state <> 'queued' OR v_job.parked_at IS NULL OR
       v_job.parked_code <> 'JOB_AUTHORITY_STALE' OR
       v_job.effect_started_at IS NOT NULL OR v_job.payload_ref IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P5689',
            MESSAGE = 'stale job is not dispositionable';
    END IF;
    IF p_action = 'adopt' AND (p_result_payload_ref IS NULL OR
       p_result_dedupe_key IS NULL OR p_result_run_at IS NULL OR
       p_result_run_at !~ 'Z$' OR p_result_run_at::timestamptz <>
           GREATEST(v_job.run_at,v_now)) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5687',
            MESSAGE = 'stale job adoption input is invalid';
    END IF;
    v_is_gap := v_job.kind = 'platform_async.event_gap_reconcile.' ||
        v_job.authority_domain AND v_job.payload_ref LIKE 'deadletter:event-gap:%';
    v_payload_ref := v_job.payload_ref;
    IF p_action = 'adopt' AND v_is_gap THEN
        IF p_result_gap_root_id IS NULL OR p_result_payload_ref IS DISTINCT FROM
            'deadletter:event-gap:' || p_result_gap_root_id::text THEN
            RAISE EXCEPTION USING ERRCODE = 'P5687',
                MESSAGE = 'stale event gap replacement identity is invalid';
        END IF;
        v_payload_ref := platform.platform_async_reissue_stale_event_gap_v1(
            v_job,p_result_gap_root_id,p_group_generation,p_domain_generation,v_now);
    ELSIF p_action = 'adopt' AND (p_result_gap_root_id IS NOT NULL OR
            p_result_payload_ref IS DISTINCT FROM v_job.payload_ref) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5687',
            MESSAGE = 'stale job replacement payload is invalid';
    END IF;
    PERFORM platform.platform_async_record_stale_job_disposition_v1(
        v_job,p_root_id,p_action_id,p_result_job_id,p_result_gap_root_id,
        p_result_payload_ref,p_result_dedupe_key,p_group_generation,
        p_domain_generation,p_action,p_payload_canonical,p_payload_sha256,
        p_reason,p_ticket,p_actor_id,p_causation_id,p_source_run_at,
        p_result_run_at,p_occurred_at,v_now
    );
    IF p_action = 'adopt' THEN
        INSERT INTO platform.jobs (
            id,tenant_id,scope_kind,scope_id,authority_group_id,
            authority_group_generation,authority_domain,cutover_generation,
            writer_origin,kind,payload_revision,payload_ref,payload_sha256,
            dedupe_key,effect_class,effect_idempotency_key,priority,run_at,
            state,attempts,max_attempts,lease_generation,retry_of_job_id,
            created_at,updated_at
        ) VALUES (
            p_result_job_id,v_job.tenant_id,v_job.scope_kind,v_job.scope_id,
            v_job.authority_group_id,p_group_generation,v_job.authority_domain,
            p_domain_generation,'v2',v_job.kind,v_job.payload_revision,
            v_payload_ref,v_job.payload_sha256,p_result_dedupe_key,
            v_job.effect_class,v_job.effect_idempotency_key,v_job.priority,
            p_result_run_at::timestamptz,'queued',0,v_job.max_attempts,0,
            v_job.id,v_now,v_now
        ) ON CONFLICT DO NOTHING;
        SELECT result.* INTO v_result FROM platform.jobs AS result
        WHERE ROW(result.id,result.tenant_id,result.scope_kind,result.scope_id,
            result.authority_group_id,result.authority_group_generation,
            result.authority_domain,result.cutover_generation,
            result.writer_origin,result.kind,result.payload_revision,
            result.payload_ref,result.payload_sha256,result.dedupe_key,
            result.effect_class,result.effect_idempotency_key,result.priority,
            result.run_at,result.max_attempts,result.retry_of_job_id,
            result.created_at,result.updated_at)
          IS NOT DISTINCT FROM ROW(p_result_job_id,v_job.tenant_id,
            v_job.scope_kind,v_job.scope_id,v_job.authority_group_id,
            p_group_generation,v_job.authority_domain,p_domain_generation,'v2',
            v_job.kind,v_job.payload_revision,v_payload_ref,v_job.payload_sha256,
            p_result_dedupe_key,v_job.effect_class,v_job.effect_idempotency_key,
            v_job.priority,p_result_run_at::timestamptz,v_job.max_attempts,
            v_job.id,v_now,v_now)
          AND result.state = 'queued' AND result.attempts = 0
          AND result.lease_generation = 0 AND result.lease_owner_id IS NULL
          AND result.parked_at IS NULL AND result.effect_started_at IS NULL
          AND result.last_error_code IS NULL AND result.completed_at IS NULL;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE = 'P5689',
                MESSAGE = 'stale job replacement conflicts';
        END IF;
    END IF;
    UPDATE platform.jobs SET state = 'cancelled',last_error_class = 'cancelled',
        last_error_code = v_code,last_error_sha256 = v_error_sha256,
        retain_until = v_now + interval '90 days',completed_at = v_now,
        updated_at = v_now WHERE id = v_job.id;
    RETURN QUERY SELECT v_now,p_action_id,p_result_job_id,
        p_result_gap_root_id,true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_dispose_stale_job_v1(
    uuid,text,uuid,uuid,uuid,uuid,text,uuid,text,bigint,text,bigint,text,text,
    bytea,text,text,text,uuid,uuid,text,text,text
) FROM PUBLIC;
