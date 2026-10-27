-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_record_stale_job_disposition_v1(
    p_job platform.jobs,p_root_id uuid,p_action_id uuid,p_result_job_id uuid,
    p_result_root_id uuid,p_result_payload_ref text,p_result_dedupe_key text,
    p_group_generation bigint,p_domain_generation bigint,p_action text,
    p_payload_canonical bytea,p_payload_sha256 text,p_reason text,p_ticket text,
    p_actor_id uuid,p_causation_id uuid,p_source_run_at text,
    p_result_run_at text,p_occurred_at text,p_now timestamptz
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_envelope jsonb;
    v_error_sha256 text;
    v_root platform.dead_letters%ROWTYPE;
BEGIN
    v_envelope := platform.platform_async_validate_stale_job_disposition_v1(
        p_job,p_root_id,p_action_id,p_result_job_id,p_result_root_id,
        p_result_payload_ref,p_result_dedupe_key,p_group_generation,
        p_domain_generation,p_action,p_payload_canonical,p_payload_sha256,
        p_reason,p_ticket,p_actor_id,p_causation_id,p_source_run_at,
        p_result_run_at,p_occurred_at,p_now
    );
    v_error_sha256 := pg_catalog.encode(pg_catalog.sha256(
        pg_catalog.convert_to('JOB_AUTHORITY_STALE','UTF8')),'hex');
    INSERT INTO platform.dead_letters (
        id,root_id,entry_sequence,record_kind,tenant_id,scope_kind,scope_id,
        authority_group_id,authority_group_generation,authority_domain,
        cutover_generation,writer_origin,source_kind,source_id,source_contract,
        source_version,source_sha256,reason_class,reason_code,
        redacted_error_sha256,payload_canonical,payload_sha256,
        redacted_envelope,full_payload_expires_at,metadata_expires_at,recorded_at
    ) VALUES (
        p_root_id,p_root_id,0,'root',p_job.tenant_id,p_job.scope_kind,
        p_job.scope_id,p_job.authority_group_id,p_group_generation,
        p_job.authority_domain,p_domain_generation,'v2','job',p_job.id,
        p_job.kind,p_job.payload_revision,p_job.payload_sha256,
        'stale_generation','JOB_AUTHORITY_STALE',v_error_sha256,
        p_payload_canonical,p_payload_sha256,v_envelope,
        p_now + interval '30 days',p_now + interval '365 days',p_now
    ) ON CONFLICT (id) DO NOTHING;
    SELECT root.* INTO v_root FROM platform.dead_letters AS root
    WHERE ROW(root.id,root.root_id,root.entry_sequence,root.record_kind,
        root.tenant_id,root.scope_kind,root.scope_id,root.authority_group_id,
        root.authority_group_generation,root.authority_domain,
        root.cutover_generation,root.writer_origin,root.source_kind,
        root.source_id,root.consumer_name,root.consumer_version,
        root.source_contract,root.source_version,root.source_sha256,
        root.reason_class,root.reason_code,root.redacted_error_sha256,
        root.payload_canonical,root.payload_sha256,root.payload_redacted_at,
        root.redacted_envelope,root.action_reason,root.action_ticket,
        root.action_actor_id,root.action_causation_id)
      IS NOT DISTINCT FROM ROW(p_root_id,p_root_id,0,'root',p_job.tenant_id,
        p_job.scope_kind,p_job.scope_id,p_job.authority_group_id,
        p_group_generation,p_job.authority_domain,p_domain_generation,'v2',
        'job',p_job.id,NULL,NULL,p_job.kind,p_job.payload_revision,
        p_job.payload_sha256,'stale_generation','JOB_AUTHORITY_STALE',
        v_error_sha256,p_payload_canonical,p_payload_sha256,NULL,v_envelope,
        NULL,NULL,NULL,NULL)
      AND root.full_payload_expires_at = root.recorded_at + interval '30 days'
      AND root.metadata_expires_at = root.recorded_at + interval '365 days'
      AND root.recorded_at <= p_now;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5685',
            MESSAGE = 'stale job disposition root conflicts';
    END IF;
    INSERT INTO platform.dead_letters (
        id,root_id,entry_sequence,record_kind,tenant_id,scope_kind,scope_id,
        authority_group_id,authority_group_generation,authority_domain,
        cutover_generation,writer_origin,source_kind,source_id,source_contract,
        source_version,source_sha256,reason_class,reason_code,
        redacted_error_sha256,payload_sha256,redacted_envelope,
        action_reason,action_ticket,action_actor_id,action_causation_id,
        full_payload_expires_at,metadata_expires_at,recorded_at
    ) VALUES (
        p_action_id,p_root_id,1,'repair_action',v_root.tenant_id,
        v_root.scope_kind,v_root.scope_id,v_root.authority_group_id,
        v_root.authority_group_generation,v_root.authority_domain,
        v_root.cutover_generation,'v2','job',v_root.source_id,
        v_root.source_contract,v_root.source_version,v_root.source_sha256,
        v_root.reason_class,v_root.reason_code,v_root.redacted_error_sha256,
        v_root.payload_sha256,v_root.redacted_envelope,p_reason,p_ticket,
        p_actor_id,p_causation_id,v_root.full_payload_expires_at,
        v_root.metadata_expires_at,p_now
    ) ON CONFLICT (id) DO NOTHING;
    IF NOT EXISTS (SELECT 1 FROM platform.dead_letters AS action WHERE
        ROW(action.id,action.root_id,action.entry_sequence,action.record_kind,
        action.tenant_id,action.scope_kind,action.scope_id,
        action.authority_group_id,action.authority_group_generation,
        action.authority_domain,action.cutover_generation,action.writer_origin,
        action.source_kind,action.source_id,action.consumer_name,
        action.consumer_version,action.source_contract,action.source_version,
        action.source_sha256,action.reason_class,action.reason_code,
        action.redacted_error_sha256,action.payload_canonical,
        action.payload_sha256,action.payload_redacted_at,
        action.redacted_envelope,action.action_reason,action.action_ticket,
        action.action_actor_id,action.action_causation_id,
        action.full_payload_expires_at,action.metadata_expires_at)
      IS NOT DISTINCT FROM ROW(p_action_id,p_root_id,1,'repair_action',
        v_root.tenant_id,v_root.scope_kind,v_root.scope_id,
        v_root.authority_group_id,v_root.authority_group_generation,
        v_root.authority_domain,v_root.cutover_generation,'v2','job',
        v_root.source_id,NULL,NULL,v_root.source_contract,v_root.source_version,
        v_root.source_sha256,v_root.reason_class,v_root.reason_code,
        v_root.redacted_error_sha256,NULL,v_root.payload_sha256,NULL,
        v_root.redacted_envelope,p_reason,p_ticket,p_actor_id,p_causation_id,
        v_root.full_payload_expires_at,v_root.metadata_expires_at)
      AND action.recorded_at BETWEEN v_root.recorded_at AND p_now) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5686',
            MESSAGE = 'stale job disposition action conflicts';
    END IF;
    RETURN p_action_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_record_stale_job_disposition_v1(
    platform.jobs,uuid,uuid,uuid,uuid,text,text,bigint,bigint,text,bytea,text,
    text,text,uuid,uuid,text,text,text,timestamptz
) FROM PUBLIC;
