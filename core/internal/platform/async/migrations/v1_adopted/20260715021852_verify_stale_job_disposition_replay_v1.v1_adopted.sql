-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_verify_stale_job_disposition_replay_v1(
    p_job platform.jobs,p_root_id uuid,p_action_id uuid,p_result_job_id uuid,
    p_group_generation bigint,p_domain_generation bigint,p_action text,
    p_reason text,p_ticket text,p_actor_id uuid,p_causation_id uuid
)
RETURNS TABLE(recorded_at timestamptz,result_root_id uuid,
    result_payload_ref text,result_dedupe_key text,result_run_at text)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_root platform.dead_letters%ROWTYPE;
    v_envelope jsonb;
    v_result platform.jobs%ROWTYPE;
    v_result_root_text text;
    v_result_run_at text;
    v_code text;
    v_error_sha256 text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_job.state <> 'cancelled' OR p_root_id IS NULL OR p_action_id IS NULL OR
       p_action NOT IN ('adopt','cancel') OR
       NOT ((p_action = 'adopt' AND p_result_job_id IS NOT NULL) OR
            (p_action = 'cancel' AND p_result_job_id IS NULL)) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5689',
            MESSAGE = 'stale job disposition replay conflicts';
    END IF;
    SELECT root.* INTO v_root FROM platform.dead_letters AS root
    WHERE root.id = p_root_id AND root.root_id = root.id
      AND root.entry_sequence = 0 AND root.record_kind = 'root'
      AND root.source_kind = 'job' AND root.source_id = p_job.id
      AND ROW(root.tenant_id,root.scope_kind,root.scope_id,
          root.authority_group_id,root.authority_group_generation,
          root.authority_domain,root.cutover_generation,root.writer_origin,
          root.source_contract,root.source_version,root.source_sha256,
          root.reason_class,root.reason_code)
          IS NOT DISTINCT FROM ROW(p_job.tenant_id,p_job.scope_kind,p_job.scope_id,
          p_job.authority_group_id,p_group_generation,p_job.authority_domain,
          p_domain_generation,'v2',p_job.kind,p_job.payload_revision,
          p_job.payload_sha256,'stale_generation','JOB_AUTHORITY_STALE');
    IF NOT FOUND OR NOT EXISTS (
        SELECT 1 FROM platform.dead_letters AS action
        WHERE action.id = p_action_id AND action.root_id = v_root.id
          AND action.entry_sequence = 1 AND action.record_kind = 'repair_action'
          AND action.action_reason = p_reason AND action.action_ticket = p_ticket
          AND action.action_actor_id = p_actor_id
          AND action.action_causation_id = p_causation_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5689',
            MESSAGE = 'stale job disposition replay evidence conflicts';
    END IF;
    v_envelope := v_root.redacted_envelope;
    IF NOT v_envelope @> pg_catalog.jsonb_build_object(
        'contract_version','stale-job-disposition.v1',
        'source_job_id',p_job.id,'decision',p_action,
        'result_job_id',p_result_job_id,'kind',p_job.kind,
        'effect_class',p_job.effect_class,
        'maximum_attempts',p_job.max_attempts,'priority',p_job.priority,
        'source_payload_sha256',p_job.payload_sha256,
        'source_dedupe_key',p_job.dedupe_key,
        'effect_idempotency_key',p_job.effect_idempotency_key,
        'current_group_generation',p_group_generation,
        'current_domain_generation',p_domain_generation
    ) OR v_envelope->>'source_run_at' IS NULL OR
        v_envelope->>'source_run_at' !~ 'Z$' OR
        (v_envelope->>'source_run_at')::timestamptz <> p_job.run_at OR
        (p_job.payload_ref IS NOT NULL AND
         v_envelope->>'source_payload_reference' <> p_job.payload_ref) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5689',
            MESSAGE = 'stale job disposition replay envelope conflicts';
    END IF;
    v_code := CASE p_action WHEN 'adopt' THEN
        'JOB_AUTHORITY_STALE_ADOPTED' ELSE 'JOB_AUTHORITY_STALE_CANCELLED' END;
    v_error_sha256 := pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        v_code || E'\n' || p_reason || E'\n' || p_ticket,'UTF8')),'hex');
    IF p_job.last_error_code IS DISTINCT FROM v_code OR
       p_job.last_error_sha256 IS DISTINCT FROM v_error_sha256 OR
       p_job.completed_at IS DISTINCT FROM v_root.recorded_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P5689',
            MESSAGE = 'stale job disposition replay outcome conflicts';
    END IF;
    IF p_action = 'cancel' THEN
        RETURN QUERY SELECT v_root.recorded_at,NULL::uuid,NULL::text,
            NULL::text,NULL::text;
        RETURN;
    END IF;
    v_result_root_text := v_envelope->>'result_root_id';
    v_result_run_at := v_envelope->>'result_run_at';
    IF v_envelope->>'result_payload_reference' IS NULL OR
       v_envelope->>'result_dedupe_key' IS NULL OR
       v_result_run_at IS NULL OR v_result_run_at !~ 'Z$' OR
       (v_result_root_text IS NOT NULL AND v_result_root_text !~
           '^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$') THEN
        RAISE EXCEPTION USING ERRCODE = 'P5689',
            MESSAGE = 'stale job adoption replay envelope conflicts';
    END IF;
    SELECT result.* INTO v_result FROM platform.jobs AS result
    WHERE ROW(result.id,result.tenant_id,result.scope_kind,result.scope_id,
        result.authority_group_id,result.authority_group_generation,
        result.authority_domain,result.cutover_generation,result.writer_origin,
        result.kind,result.payload_revision,result.payload_sha256,
        result.dedupe_key,result.effect_class,result.effect_idempotency_key,
        result.priority,result.run_at,result.max_attempts,result.retry_of_job_id,
        result.created_at) IS NOT DISTINCT FROM ROW(p_result_job_id,p_job.tenant_id,
        p_job.scope_kind,p_job.scope_id,p_job.authority_group_id,
        p_group_generation,p_job.authority_domain,p_domain_generation,'v2',
        p_job.kind,p_job.payload_revision,p_job.payload_sha256,
        v_envelope->>'result_dedupe_key',p_job.effect_class,
        p_job.effect_idempotency_key,p_job.priority,
        v_result_run_at::timestamptz,p_job.max_attempts,p_job.id,
        v_root.recorded_at)
      AND (result.payload_ref = v_envelope->>'result_payload_reference' OR
           (result.payload_ref IS NULL AND result.payload_redacted_at IS NOT NULL));
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5689',
            MESSAGE = 'stale job adoption replay conflicts';
    END IF;
    RETURN QUERY SELECT v_root.recorded_at,v_result_root_text::uuid,
        v_envelope->>'result_payload_reference',
        v_envelope->>'result_dedupe_key',v_result_run_at;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_verify_stale_job_disposition_replay_v1(
    platform.jobs,uuid,uuid,uuid,bigint,bigint,text,text,text,uuid,uuid
) FROM PUBLIC;
