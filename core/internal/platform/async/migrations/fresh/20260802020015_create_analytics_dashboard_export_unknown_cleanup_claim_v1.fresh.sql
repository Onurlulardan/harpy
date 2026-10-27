-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_unknown_cleanup_claim_v1(
    p_tenant_id uuid,p_export_id uuid,p_job_id uuid,p_attempt_id uuid,
    p_attempt_generation bigint,p_original_lease_owner_id uuid,p_effect_id uuid,
    p_request_sha256 bytea,p_cleanup_lease_owner_id uuid,
    p_observed_at timestamptz
)
RETURNS TABLE(claim_state text,cleanup_generation bigint,lease_until timestamptz)
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_claim platform.analytics_dashboard_export_unknown_cleanup_claims%ROWTYPE;
    v_request text;
    v_claimed_at timestamptz;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export cleanup claim authority is invalid';
    END IF;
    IF p_tenant_id IS NULL OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR
       pg_catalog.uuid_extract_version(p_original_lease_owner_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 7 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_request_sha256=pg_catalog.decode(pg_catalog.repeat('00',32),'hex') OR
       pg_catalog.uuid_extract_version(p_cleanup_lease_owner_id) IS DISTINCT FROM 7 OR
       NOT pg_catalog.isfinite(p_observed_at) OR
       p_observed_at>pg_catalog.clock_timestamp()+interval '5 seconds' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics dashboard export cleanup claim input is invalid';
    END IF;
    v_request:=pg_catalog.encode(p_request_sha256,'hex');
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
      AND attempt.lease_owner_id=p_original_lease_owner_id FOR SHARE;
    IF v_job.state IS DISTINCT FROM 'dead' OR
       v_job.last_error_class IS NOT NULL OR
       v_job.last_error_code IS DISTINCT FROM 'JOB_OUTCOME_UNKNOWN' OR
       v_attempt.outcome IS DISTINCT FROM 'dead' OR
       v_attempt.error_class IS NOT NULL OR
       v_attempt.error_code IS DISTINCT FROM 'JOB_OUTCOME_UNKNOWN' OR
       v_attempt.effect_id IS DISTINCT FROM p_effect_id OR
       v_attempt.effect_request_sha256 IS DISTINCT FROM v_request OR
       v_attempt.effect_started_at IS NULL OR v_attempt.effect_receipt_id IS NOT NULL OR
       v_attempt.finished_at IS DISTINCT FROM v_job.completed_at THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export cleanup evidence differs';
    END IF;
    PERFORM platform.platform_async_analytics_dashboard_export_assert_cleanup_owner_v1(
        v_job.scope_kind,v_job.scope_id,v_job.authority_group_id,
        v_job.authority_domain);
    IF EXISTS (
        SELECT 1 FROM platform.analytics_dashboard_export_unknown_resolutions resolution
        WHERE resolution.job_id=p_job_id
    ) THEN
        claim_state:='resolved'; RETURN NEXT; RETURN;
    END IF;
    SELECT claim.* INTO v_claim
    FROM platform.analytics_dashboard_export_unknown_cleanup_claims AS claim
    WHERE claim.job_id=p_job_id FOR UPDATE;
    IF FOUND AND (v_claim.tenant_id<>p_tenant_id OR
       v_claim.export_id<>p_export_id OR v_claim.attempt_id<>p_attempt_id OR
       v_claim.attempt_generation<>p_attempt_generation OR
       v_claim.original_lease_owner_id<>p_original_lease_owner_id OR
       v_claim.effect_id<>p_effect_id OR v_claim.request_sha256<>v_request) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export cleanup claim replay differs';
    END IF;
    IF FOUND AND v_claim.lease_until>pg_catalog.clock_timestamp() THEN
        IF v_claim.lease_owner_id=p_cleanup_lease_owner_id THEN
            claim_state:='claimed'; cleanup_generation:=v_claim.claim_generation;
            lease_until:=v_claim.lease_until; RETURN NEXT; RETURN;
        END IF;
        claim_state:='busy'; RETURN NEXT; RETURN;
    END IF;
    IF FOUND AND v_claim.run_at>pg_catalog.clock_timestamp() THEN
        claim_state:='busy'; RETURN NEXT; RETURN;
    END IF;
    v_claimed_at:=GREATEST(pg_catalog.clock_timestamp(),p_observed_at,
        COALESCE(v_claim.updated_at,p_observed_at)+interval '1 microsecond');
    IF FOUND THEN
        UPDATE platform.analytics_dashboard_export_unknown_cleanup_claims
        SET claim_generation=claim_generation+1,
            lease_owner_id=p_cleanup_lease_owner_id,
            lease_heartbeat_at=v_claimed_at,
            lease_until=v_claimed_at+interval '60 seconds',
            updated_at=v_claimed_at WHERE job_id=p_job_id
        RETURNING claim_generation INTO cleanup_generation;
    ELSE
        cleanup_generation:=1;
        INSERT INTO platform.analytics_dashboard_export_unknown_cleanup_claims(
            job_id,tenant_id,export_id,attempt_id,attempt_generation,
            original_lease_owner_id,effect_id,request_sha256,claim_generation,
            lease_owner_id,lease_heartbeat_at,lease_until,run_at,created_at,updated_at
        ) VALUES (p_job_id,p_tenant_id,p_export_id,p_attempt_id,p_attempt_generation,
            p_original_lease_owner_id,p_effect_id,v_request,cleanup_generation,
            p_cleanup_lease_owner_id,v_claimed_at,v_claimed_at+interval '60 seconds',
            v_claimed_at,v_claimed_at,v_claimed_at);
    END IF;
    claim_state:='claimed'; lease_until:=v_claimed_at+interval '60 seconds';
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_unknown_cleanup_claim_v1(
        uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,uuid,timestamptz)
FROM PUBLIC;
