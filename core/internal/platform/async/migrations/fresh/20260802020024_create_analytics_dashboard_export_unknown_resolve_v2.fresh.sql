-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_unknown_resolve_v2(
    p_tenant_id uuid,p_export_id uuid,p_job_id uuid,p_attempt_id uuid,
    p_attempt_generation bigint,p_lease_owner_id uuid,p_effect_id uuid,
    p_request_sha256 bytea,p_cleanup_generation bigint,
    p_cleanup_lease_owner_id uuid,p_resolution_receipt_id uuid,
    p_resolution_kind text,p_resolution_sha256 bytea,p_resolved_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_claim platform.analytics_dashboard_export_unknown_cleanup_claims%ROWTYPE;
    v_resolution platform.analytics_dashboard_export_unknown_resolutions%ROWTYPE;
    v_request text;
    v_resolution_digest text;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export unknown resolution authority is invalid';
    END IF;
    IF p_tenant_id IS NULL OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR
       pg_catalog.uuid_extract_version(p_lease_owner_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 7 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_request_sha256=pg_catalog.decode(pg_catalog.repeat('00',32),'hex') OR
       p_cleanup_generation<1 OR
       pg_catalog.uuid_extract_version(p_cleanup_lease_owner_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_resolution_receipt_id) IS DISTINCT FROM 7 OR
       p_resolution_receipt_id IN (p_job_id,p_attempt_id,p_effect_id) OR
       p_resolution_kind NOT IN ('deleted','absent') OR
       pg_catalog.octet_length(p_resolution_sha256) IS DISTINCT FROM 32 OR
       p_resolution_sha256=pg_catalog.decode(pg_catalog.repeat('00',32),'hex') OR
       NOT pg_catalog.isfinite(p_resolved_at) OR
       p_resolved_at>pg_catalog.clock_timestamp()+interval '5 seconds' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics dashboard export unknown resolution input is invalid';
    END IF;
    v_request:=pg_catalog.encode(p_request_sha256,'hex');
    v_resolution_digest:=pg_catalog.encode(p_resolution_sha256,'hex');
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id=p_job_id AND job.tenant_id=p_tenant_id
      AND job.scope_kind='tenant' AND job.scope_id=p_tenant_id
      AND job.authority_domain='analytics' AND job.writer_origin='v2'
      AND job.kind='analytics.dashboard.export.render'
      AND job.payload_revision=1 AND job.effect_class='external'
      AND job.max_attempts=3 AND job.payload_ref=p_export_id::text
      AND job.payload_sha256=pg_catalog.encode(
          pg_catalog.sha256(pg_catalog.convert_to(p_export_id::text,'UTF8')),'hex')
      AND job.dedupe_key='analytics.dashboard.export.render:'||p_export_id::text
      AND job.effect_idempotency_key=job.dedupe_key FOR UPDATE;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.attempt_id=p_attempt_id AND attempt.job_id=p_job_id
      AND attempt.scope_kind='tenant' AND attempt.scope_id=p_tenant_id
      AND attempt.attempt=p_attempt_generation
      AND attempt.lease_generation=p_attempt_generation
      AND attempt.lease_owner_id=p_lease_owner_id FOR SHARE;
    IF v_job.id IS NULL OR v_attempt.attempt_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export unknown evidence is unavailable';
    END IF;
    PERFORM platform.platform_async_analytics_dashboard_export_assert_cleanup_owner_v1(
        v_job.scope_kind,v_job.scope_id,v_job.authority_group_id,
        v_job.authority_domain);
    IF NOT COALESCE((v_job.state='dead' AND
       v_job.last_error_class IS NULL AND
       v_job.last_error_code='JOB_OUTCOME_UNKNOWN' AND
       v_job.last_error_sha256 IS NULL AND v_job.completed_at IS NOT NULL AND
       v_job.updated_at=v_job.completed_at AND p_resolved_at>v_job.completed_at AND
       v_job.lease_owner_id IS NULL AND v_job.lease_heartbeat_at IS NULL AND
       v_job.lease_until IS NULL AND
       v_job.retain_until=v_job.completed_at+interval '366 days' AND
       v_attempt.finished_at=v_job.completed_at AND
       v_attempt.outcome='dead' AND v_attempt.error_class IS NULL AND
       v_attempt.error_code='JOB_OUTCOME_UNKNOWN' AND
       v_attempt.redacted_error_sha256 IS NULL AND
       v_attempt.effect_receipt_id IS NULL AND
       v_attempt.effect_id=p_effect_id AND
       v_attempt.effect_request_sha256=v_request AND
       v_attempt.effect_started_at IS NOT NULL AND
       v_job.effect_started_at=v_attempt.effect_started_at),false) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export is not exact unknown evidence';
    END IF;
    SELECT resolution.* INTO v_resolution
    FROM platform.analytics_dashboard_export_unknown_resolutions AS resolution
    WHERE resolution.job_id=p_job_id FOR SHARE;
    IF FOUND THEN
        IF v_resolution.resolution_receipt_id=p_resolution_receipt_id AND
           v_resolution.tenant_id=p_tenant_id AND
           v_resolution.export_id=p_export_id AND
           v_resolution.attempt_id=p_attempt_id AND
           v_resolution.attempt_generation=p_attempt_generation AND
           v_resolution.lease_owner_id=p_lease_owner_id AND
           v_resolution.effect_id=p_effect_id AND
           v_resolution.request_sha256=v_request AND
           v_resolution.cleanup_generation=p_cleanup_generation AND
           v_resolution.cleanup_lease_owner_id=p_cleanup_lease_owner_id AND
           v_resolution.resolution_kind=p_resolution_kind AND
           v_resolution.resolution_sha256=v_resolution_digest AND
           v_resolution.resolved_at=p_resolved_at THEN
            RETURN true;
        END IF;
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export unknown resolution replay differs';
    END IF;
    SELECT claim.* INTO v_claim
    FROM platform.analytics_dashboard_export_unknown_cleanup_claims AS claim
    WHERE claim.job_id=p_job_id AND claim.tenant_id=p_tenant_id
      AND claim.export_id=p_export_id AND claim.attempt_id=p_attempt_id
      AND claim.attempt_generation=p_attempt_generation
      AND claim.original_lease_owner_id=p_lease_owner_id
      AND claim.effect_id=p_effect_id AND claim.request_sha256=v_request
    FOR UPDATE;
    IF v_claim.job_id IS NULL OR
       v_claim.claim_generation<>p_cleanup_generation OR
       v_claim.lease_owner_id<>p_cleanup_lease_owner_id OR
       v_claim.lease_until<=pg_catalog.clock_timestamp() THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export cleanup resolution lease is stale';
    END IF;
    INSERT INTO platform.analytics_dashboard_export_unknown_resolutions(
        resolution_receipt_id,tenant_id,export_id,job_id,attempt_id,
        attempt_generation,lease_owner_id,effect_id,request_sha256,
        cleanup_generation,cleanup_lease_owner_id,resolution_kind,
        resolution_sha256,resolved_at
    ) VALUES (
        p_resolution_receipt_id,p_tenant_id,p_export_id,p_job_id,p_attempt_id,
        p_attempt_generation,p_lease_owner_id,p_effect_id,v_request,
        p_cleanup_generation,p_cleanup_lease_owner_id,p_resolution_kind,
        v_resolution_digest,p_resolved_at
    );
    UPDATE platform.artifact_quota_counters
    SET reserved_ciphertext_bytes=reserved_ciphertext_bytes-2147483648,
        revision=revision+1,updated_at=GREATEST(
            updated_at+interval '1 microsecond',p_resolved_at)
    WHERE tenant_id=p_tenant_id AND reserved_ciphertext_bytes>=2147483648;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6049',
            MESSAGE='Analytics dashboard export unknown quota is inconsistent';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_unknown_resolve_v2(
        uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,bigint,uuid,uuid,text,bytea,
        timestamptz)
FROM PUBLIC;
