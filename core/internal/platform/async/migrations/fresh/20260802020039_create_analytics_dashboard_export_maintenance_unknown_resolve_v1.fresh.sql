-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_maintenance_unknown_resolve_v1(
    p_kind text,p_tenant_id uuid,p_tenant_schema name,p_export_id uuid,
    p_authority_group_id text,p_authority_group_generation bigint,
    p_cutover_generation bigint,p_job_id uuid,p_attempt_id uuid,
    p_attempt_number integer,p_lease_generation bigint,
    p_original_lease_owner_id uuid,p_effect_id uuid,p_request_sha256 bytea,
    p_unknown_at timestamptz,p_cleanup_generation bigint,p_cleanup_lease_owner_id uuid,
    p_resolution_receipt_id uuid,p_effect_receipt_id uuid,p_effect_receipt_sha256 bytea,
    p_terminal_state text,p_resolution_sha256 bytea,p_resolved_at timestamptz
)
RETURNS TABLE(replayed boolean,effect_receipt_id uuid,job_state text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_claim platform.analytics_dashboard_export_maintenance_unknown_claims%ROWTYPE;
    v_resolution platform.analytics_dashboard_export_maintenance_unknown_resolutions%ROWTYPE;
    v_schema name;
    v_request text;
    v_receipt_digest text;
    v_resolution_digest text;
    v_code text;
    v_result_ref text;
    v_result_sha text;
    v_max integer;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export maintenance resolution authority is invalid';
    END IF;
    v_max:=CASE p_kind WHEN 'analytics.dashboard.export.delete' THEN 3
        WHEN 'analytics.dashboard.export.expire' THEN 8 END;
    IF v_max IS NULL OR p_tenant_id IS NULL OR
       (p_tenant_schema::text~'^[a-z][a-z0-9_]{0,62}$') IS DISTINCT FROM true OR
       (p_authority_group_id~'^[a-z][a-z0-9_]{0,62}$') IS DISTINCT FROM true OR
       COALESCE(p_authority_group_generation<1,true) OR
       COALESCE(p_cutover_generation<1,true) OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       p_attempt_number NOT BETWEEN 1 AND v_max OR COALESCE(p_lease_generation<1,true) OR
       pg_catalog.uuid_extract_version(p_original_lease_owner_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 5 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_request_sha256=pg_catalog.decode(pg_catalog.repeat('00',32),'hex') OR
       pg_catalog.isfinite(p_unknown_at) IS DISTINCT FROM true OR
       COALESCE(p_cleanup_generation<1,true) OR
       pg_catalog.uuid_extract_version(p_cleanup_lease_owner_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_resolution_receipt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_effect_receipt_id) IS DISTINCT FROM 5 OR
       p_effect_receipt_id=p_effect_id OR
       pg_catalog.octet_length(p_effect_receipt_sha256) IS DISTINCT FROM 32 OR
       p_effect_receipt_sha256=pg_catalog.decode(pg_catalog.repeat('00',32),'hex') OR
       pg_catalog.octet_length(p_resolution_sha256) IS DISTINCT FROM 32 OR
       p_resolution_sha256=pg_catalog.decode(pg_catalog.repeat('00',32),'hex') OR
       pg_catalog.isfinite(p_resolved_at) IS DISTINCT FROM true OR
       p_unknown_at>=p_resolved_at OR
       p_resolved_at>pg_catalog.clock_timestamp()+interval '5 seconds' OR
       (CASE p_kind WHEN 'analytics.dashboard.export.delete' THEN
            p_terminal_state='deleted' ELSE p_terminal_state IN
            ('expired','cancelled','failed','deleted') END) IS DISTINCT FROM true THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics dashboard export maintenance resolution input is invalid';
    END IF;
    v_request:=pg_catalog.encode(p_request_sha256,'hex');
    v_receipt_digest:=pg_catalog.encode(p_effect_receipt_sha256,'hex');
    v_resolution_digest:=pg_catalog.encode(p_resolution_sha256,'hex');
    SELECT job.* INTO v_job FROM platform.jobs job
    WHERE job.id=p_job_id AND job.tenant_id=p_tenant_id AND job.scope_kind='tenant'
      AND job.scope_id=p_tenant_id AND job.authority_group_id=p_authority_group_id
      AND job.authority_group_generation=p_authority_group_generation
      AND job.authority_domain='analytics' AND job.cutover_generation=p_cutover_generation
      AND job.writer_origin='v2' AND job.kind=p_kind AND job.payload_revision=1
      AND job.effect_class='external' AND job.max_attempts=v_max
      AND job.payload_ref=p_export_id::text
      AND job.payload_sha256=pg_catalog.encode(pg_catalog.sha256(
          pg_catalog.convert_to(p_export_id::text,'UTF8')),'hex')
      AND job.dedupe_key=p_kind||':'||p_export_id::text
      AND job.effect_idempotency_key=job.dedupe_key FOR UPDATE;
    IF v_job.id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance resolution job differs';
    END IF;
    SELECT tenant.schema_name INTO v_schema FROM platform.tenants tenant
    WHERE tenant.id=p_tenant_id AND tenant.deleted_at IS NULL FOR SHARE;
    IF v_schema IS DISTINCT FROM p_tenant_schema THEN
        RAISE EXCEPTION USING ERRCODE='P6046',
            MESSAGE='Analytics dashboard export maintenance tenant schema changed';
    END IF;
    PERFORM platform.platform_async_analytics_dashboard_export_assert_cleanup_owner_v1(
        v_job.scope_kind,v_job.scope_id,v_job.authority_group_id,v_job.authority_domain);
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts attempt
    WHERE attempt.attempt_id=p_attempt_id AND attempt.job_id=p_job_id
      AND attempt.scope_kind='tenant' AND attempt.scope_id=p_tenant_id
      AND attempt.attempt=p_attempt_number AND attempt.lease_generation=p_lease_generation
      AND attempt.lease_owner_id=p_original_lease_owner_id FOR SHARE;
    SELECT claim.* INTO v_claim
    FROM platform.analytics_dashboard_export_maintenance_unknown_claims claim
    WHERE claim.job_id=p_job_id AND claim.kind=p_kind AND claim.tenant_id=p_tenant_id
      AND claim.tenant_schema=p_tenant_schema AND claim.export_id=p_export_id
      AND claim.authority_group_id=p_authority_group_id
      AND claim.authority_group_generation=p_authority_group_generation
      AND claim.cutover_generation=p_cutover_generation AND claim.attempt_id=p_attempt_id
      AND claim.attempt_number=p_attempt_number AND claim.lease_generation=p_lease_generation
      AND claim.original_lease_owner_id=p_original_lease_owner_id
      AND claim.effect_id=p_effect_id AND claim.request_sha256=v_request
      AND claim.unknown_at=p_unknown_at AND claim.claim_generation=p_cleanup_generation
      AND claim.cleanup_lease_owner_id=p_cleanup_lease_owner_id FOR UPDATE;
    SELECT resolution.* INTO v_resolution
    FROM platform.analytics_dashboard_export_maintenance_unknown_resolutions resolution
    WHERE resolution.job_id=p_job_id FOR SHARE;
    IF FOUND THEN
        PERFORM platform.platform_async_analytics_dashboard_export_maintenance_assert_replay_v1(
            v_job,v_attempt,v_claim,v_resolution,p_effect_receipt_id,
            p_effect_receipt_sha256,p_resolution_receipt_id,p_resolution_sha256,
            p_terminal_state,p_resolved_at);
        replayed:=true; effect_receipt_id:=p_effect_receipt_id;
        job_state:='succeeded'; RETURN NEXT; RETURN;
    END IF;
    IF v_claim.job_id IS NULL OR v_attempt.attempt_id IS NULL OR
       v_claim.lease_until<=pg_catalog.clock_timestamp() OR v_job.state<>'dead' OR
       v_job.last_error_class IS NOT NULL OR
       v_job.last_error_code<>'JOB_OUTCOME_UNKNOWN' OR
       v_job.last_error_sha256 IS NOT NULL OR v_job.completed_at<>p_unknown_at OR
       v_job.completed_at<>v_attempt.finished_at OR v_attempt.outcome<>'dead' OR
       v_attempt.error_class IS NOT NULL OR
       v_attempt.error_code<>'JOB_OUTCOME_UNKNOWN' OR
       v_attempt.redacted_error_sha256 IS NOT NULL OR v_attempt.effect_id<>p_effect_id OR
       v_attempt.effect_request_sha256<>v_request OR v_attempt.effect_started_at IS NULL OR
       v_attempt.effect_receipt_id IS NOT NULL OR EXISTS (
           SELECT 1 FROM platform.idempotency_receipts receipt
           WHERE receipt.receipt_id=p_effect_receipt_id OR
             (receipt.record_kind='effect' AND receipt.effect_id=p_effect_id AND
              receipt.effect_key=v_job.effect_idempotency_key)) OR
       platform.platform_async_analytics_dashboard_export_maintenance_receipt_sha256_v1(
           v_job,v_attempt,p_effect_receipt_id,p_terminal_state,p_resolved_at)
           <>p_effect_receipt_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance resolution evidence differs';
    END IF;
    SELECT inserted.provider_code,inserted.result_reference,inserted.result_sha256
    INTO v_code,v_result_ref,v_result_sha
    FROM platform.platform_async_analytics_dashboard_export_maintenance_insert_receipt_v1(
        v_job,v_attempt,p_effect_receipt_id,p_effect_receipt_sha256,
        p_terminal_state,p_resolved_at) inserted;
    INSERT INTO platform.analytics_dashboard_export_maintenance_unknown_resolutions(
        resolution_receipt_id,effect_receipt_id,effect_receipt_sha256,resolution_sha256,
        kind,tenant_id,tenant_schema,export_id,authority_group_id,
        authority_group_generation,cutover_generation,job_id,attempt_id,attempt_number,
        lease_generation,original_lease_owner_id,effect_id,request_sha256,
        original_unknown_at,original_attempt_finished_at,original_attempt_outcome,
        original_error_class,original_error_code,original_error_sha256,
        cleanup_generation,cleanup_lease_owner_id,terminal_state,provider_code,
        provider_reference,result_reference,result_sha256,resolved_at
    ) VALUES (p_resolution_receipt_id,p_effect_receipt_id,v_receipt_digest,
        v_resolution_digest,p_kind,p_tenant_id,p_tenant_schema,p_export_id,
        v_job.authority_group_id,v_job.authority_group_generation,
        v_job.cutover_generation,p_job_id,p_attempt_id,p_attempt_number,
        p_lease_generation,p_original_lease_owner_id,p_effect_id,v_request,
        v_job.completed_at,v_attempt.finished_at,v_attempt.outcome,
        v_attempt.error_class,v_attempt.error_code,v_attempt.redacted_error_sha256,
        p_cleanup_generation,p_cleanup_lease_owner_id,p_terminal_state,v_code,
        p_export_id::text,v_result_ref,v_result_sha,p_resolved_at);
    UPDATE platform.job_attempts SET finished_at=p_resolved_at,outcome='succeeded',
        error_code=NULL,effect_receipt_id=p_effect_receipt_id
    WHERE attempt_id=p_attempt_id AND outcome='dead'
      AND error_code='JOB_OUTCOME_UNKNOWN'
      AND platform.job_attempts.effect_receipt_id IS NULL;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE='P6048',
        MESSAGE='Analytics dashboard export maintenance attempt changed'; END IF;
    UPDATE platform.jobs SET state='succeeded',last_error_code=NULL,
        completed_at=p_resolved_at,updated_at=p_resolved_at,
        retain_until=p_resolved_at+interval '90 days'
    WHERE id=p_job_id AND state='dead' AND last_error_code='JOB_OUTCOME_UNKNOWN'
      AND completed_at=p_unknown_at;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE='P6048',
        MESSAGE='Analytics dashboard export maintenance job changed'; END IF;
    replayed:=false; effect_receipt_id:=p_effect_receipt_id; job_state:='succeeded';
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_maintenance_unknown_resolve_v1(
        text,uuid,name,uuid,text,bigint,bigint,uuid,uuid,integer,bigint,uuid,uuid,
        bytea,timestamptz,bigint,uuid,uuid,uuid,bytea,text,bytea,timestamptz)
FROM PUBLIC;
